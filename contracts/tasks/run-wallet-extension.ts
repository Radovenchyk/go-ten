import { task } from "hardhat/config";

import * as dockerApi from 'node-docker-api';

import * as url from 'node:url';
import http from 'http';


task("run-wallet-extension", "Starts up the wallet extension docker container.")
.addFlag('wait')
.addParam('dockerImage', 
    'The docker image to use for wallet extension', 
    'testnetobscuronet.azurecr.io/obscuronet/walletextension')
.addParam('rpcUrl', "Which network to pick the node connection info from?")
.setAction(async function(args, hre) {
    const docker = new dockerApi.Docker({ socketPath: '/var/run/docker.sock' });

    const parsedUrl = url.parse(args.rpcUrl)

    const container = await docker.container.create({
        Image: args.dockerImage,
        Cmd: [
            "--port=3000",
            "--portWS=3001",
            `--nodeHost=${parsedUrl.hostname}`,
            `--nodePortWS=${parsedUrl.port}`
        ],
        ExposedPorts: { "3000/tcp": {}, "3001/tcp": {}, "3000/udp": {}, "3001/udp": {} },
        PortBindings:  { "3000/tcp": [{ "HostPort": "3000" }], "3001/tcp": [{ "HostPort": "3001" }] }
    })


    process.on('SIGINT', ()=>{
        container.stop();
    })
    
    await container.start();

    const stream: any = await container.logs({
        follow: true,
        stdout: true,
        stderr: true
    })

    console.log(`\nWallet Extension{ ${container.id.slice(0, 5)} } %>\n`);
    const startupPromise = new Promise((resolveInner)=> {    
        stream.on('data', (msg: any)=> {
            const message = msg.toString();

            console.log(message);    
            if(message.includes("Wallet extension started")) {
                console.log(`Wallet - success!`);
                resolveInner(true);
            }
        });

        setTimeout(resolveInner, 40_000);
    });

    await startupPromise;
    console.log("\n[ . . . ]\n");


    if (args.wait) {   
        await container.wait();
    }
});

task("stop-wallet-extension", "Starts up the wallet extension docker container.")
.addParam('dockerImage', 
    'The docker image to use for wallet extension', 
    'testnetobscuronet.azurecr.io/obscuronet/walletextension')
.setAction(async function(args, hre) {
    const docker = new dockerApi.Docker({ socketPath: '/var/run/docker.sock' });
    const containers = await docker.container.list();

    const container = containers.find((c)=> { 
       const data : any = c.data; 
       return data.Image == 'testnetobscuronet.azurecr.io/obscuronet/walletextension'
    })

    await container?.stop()
});


task("add-key", "Creates a viewing key for a specifiec address")
.addParam("address", "The address for which to add key")
.setAction(async function(args, hre) {
    async function viewingKeyForAddress(address: string) : Promise<string> {
        return new Promise((resolve, fail)=> {
    
            const data = {"address": address}
    
            const req = http.request({
                host: 'localhost',
                port: 3000,
                path: '/generateviewingkey/',
                method: 'post',
                headers: {
                    'Content-Type': 'application/json'
                }
            }, (response)=>{
                if (response.statusCode != 200) {
                    fail(response.statusCode);
                    return;
                }
    
                let chunks : string[] = []
                response.on('data', (chunk)=>{
                    chunks.push(chunk);
                });
    
                response.on('end', ()=> { 
                    resolve(chunks.join('')); 
                });
            });
            req.write(JSON.stringify(data));
            req.end()
            setTimeout(resolve, 15_000);
        });
    }
    
    interface SignedData { signature: string, address: string }
    
    async function submitKey(signedData: SignedData) : Promise<number> {
        return await new Promise(async (resolve, fail)=>{ 
            const req = http.request({
                host: 'localhost',
                port: 3000,
                path: '/submitviewingkey/',
                method: 'post',
                headers: {
                    'Content-Type': 'application/json'
                }
            }, (response)=>{
                if (response.statusCode == 200) { 
                    resolve(response.statusCode);
                } else {
                    fail(response.statusCode);
                }
            });
    
            req.write(JSON.stringify(signedData));
            req.end()
        });
    }

    const key = await viewingKeyForAddress(args.address);

    const signaturePromise = (await hre.ethers.getSigner(args.address)).signMessage(`vk${key}`);
    const signedData = { 'signature': await signaturePromise, 'address': args.address };
    await submitKey(signedData)
});