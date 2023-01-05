import {HardhatRuntimeEnvironment} from 'hardhat/types';
import {DeployFunction} from 'hardhat-deploy/types';
import { Receipt } from 'hardhat-deploy/dist/types';

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
    const l1Network = hre.companionNetworks.layer1;
    const l2Network = hre; 

    const l1Accounts = await l1Network.getNamedAccounts();
    const l2Accounts = await l2Network.getNamedAccounts();

    console.log(`L2_003 - multi deployer`);

    const layer2BridgeDeployment = await l2Network.deployments.get("EthereumBridge");
    const HOCDeployment = await l1Network.deployments.get("HOCERC20");
    const POCDeployment = await l1Network.deployments.get("POCERC20");

    const setResult = await l1Network.deployments.execute("ObscuroBridge", {
        from: l1Accounts.deployer, 
        log: true,
    }, "setRemoteBridge", layer2BridgeDeployment.address);
    if (setResult.status != 1) {
        console.error("Unable to link L1 and L2 bridges!");
        throw Error("Unable to link L1 and L2 bridges!");
    }

    console.log(`setRemoteBridge = ${layer2BridgeDeployment.address}`);

    let hocResultPromise = l1Network.deployments.execute("ObscuroBridge", {
        from: l1Accounts.deployer, 
        log: true,
    }, "whitelistToken", HOCDeployment.address, "HOC", "HOC");

    const hocResult = (await hocResultPromise); 
    if (hocResult.status != 1) {
        console.error("Unable to whitelist HOC token!");
        throw Error("Unable to whitelist HOC token!");
    }


    const pocResultPromise = l1Network.deployments.execute("ObscuroBridge", {
        from: l1Accounts.deployer, 
        log: true,
    }, "whitelistToken", POCDeployment.address, "POC", "POC");
    
    const pocResult = (await pocResultPromise);
    if (pocResult.status != 1) {
        console.error("Unable to whitelist POC token!");
        throw Error("Unable to whitelist POC token!");
    }

    const eventSignature = "LogMessagePublished(address,uint64,uint32,uint32,bytes,uint8)";
    const topic = hre.ethers.utils.id(eventSignature)
    let eventIface = new hre.ethers.utils.Interface([ `event ${eventSignature}`]);

    function getXChainMessages(result: Receipt) {
        const events = result.logs?.filter((x)=> { 
            return x.topics.find((t: string)=> t == topic) != undefined;
        });

        const messages = events!.map((event)=> {
            const decodedEvent = eventIface.parseLog({
                topics: event!.topics!,
                data: event!.data
            });
        
            const xchainMessage = {
                sender: decodedEvent.args[0],
                sequence: decodedEvent.args[1],
                nonce: decodedEvent.args[2],
                topic: decodedEvent.args[3],
                payload: decodedEvent.args[4],
                consistencyLevel: decodedEvent.args[5]
            };

            return xchainMessage;
        })

        return messages;
    }

    let messages = getXChainMessages(hocResult);
    messages = messages.concat(getXChainMessages(pocResult));

    // Freeze until the enclave processes the blocks and picks up the messages that have been carried over.
    await new Promise(resolve=>setTimeout(resolve, 2_000));
    const relayMsg = async (msg: any) => {
        return l2Network.deployments.execute("CrossChainMessenger", {
            from: l2Accounts.deployer, 
            log: true,
        }, "relayMessage", msg);
    };

    const hocRelayRes = await relayMsg(messages[0]);
    const pocRelayRes = await relayMsg(messages[1]);

    [ hocRelayRes, pocRelayRes ].forEach(res=>{
        if (res.status != 1) {
            throw Error("Unable to relay messages...");
        } 
    });
};

export default func;
func.tags = ['Whitelist', 'Whitelist_deploy'];
func.dependencies = ['EthereumBridge'];
