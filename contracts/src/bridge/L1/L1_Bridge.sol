// SPDX-License-Identifier: Apache 2

pragma solidity >=0.7.0 <0.9.0;

import "./IL1Bridge.sol";
import "../IBridge.sol";
import "../IBridgeSubordinate.sol";
import "../../messaging/messenger/CrossChainEnabledObscuro.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract ObscuroBridge is CrossChainEnabledObscuro, IBridge, IL1Bridge, AccessControl  {

    bytes32 public constant NATIVE_TOKEN_ROLE = keccak256("NATIVE_TOKEN");
    bytes32 public constant ERC20_TOKEN_ROLE = keccak256("ERC20_TOKEN");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    address remoteBridgeAddress;

    constructor(address messenger)
    CrossChainEnabledObscuro(messenger)
    {
        _grantRole(ADMIN_ROLE, msg.sender);
        _grantRole(NATIVE_TOKEN_ROLE, address(0x0));
    }

    function whitelistToken(address asset, string calldata name, string calldata symbol) external onlyRole(ADMIN_ROLE) {
        _grantRole(ERC20_TOKEN_ROLE, asset);

        bytes memory data = abi.encodeWithSelector(IBridgeSubordinate.createWrappedToken.selector, asset, name, symbol);
        queueMessage(remoteBridgeAddress, data, uint32(Topics.MANAGEMENT), 0, 0);
    } 

    function removeToken(address asset) external onlyRole(ADMIN_ROLE) {
        _revokeRole(ERC20_TOKEN_ROLE, asset);
    }

    function setRemoteBridge(address bridge) external onlyRole(ADMIN_ROLE) {
        remoteBridgeAddress = bridge;
    }

    function sendNative(address receiver) override external payable {
        require(msg.value > 0, "Empty transfer.");

        bytes memory data = abi.encodeWithSelector(IBridge.receiveAssets.selector, address(0x0), msg.value, receiver);
        queueMessage(remoteBridgeAddress, data, uint32(Topics.TRANSFER), 0, 0);
    }

    function sendAssets(address asset, uint256 amount, address receiver) override external {
        require(amount > 0, "Attempting empty transfer.");
        require(hasRole(ERC20_TOKEN_ROLE, asset), "This address has not been given a type and is thus considered not whitelisted.");

        SafeERC20.safeTransferFrom(IERC20(asset), msg.sender, address(this), amount);
        
        bytes memory data = abi.encodeWithSelector(IBridge.receiveAssets.selector, asset, amount, receiver);
        queueMessage(remoteBridgeAddress, data, uint32(Topics.TRANSFER), 0, 0);
    }

    function receiveAssets(address asset, uint256 amount, address receiver) 
    override 
    external 
    onlyCrossChainSender(remoteBridgeAddress) {
        if (hasRole(ERC20_TOKEN_ROLE, asset)) {
            _receiveTokens(asset, amount, receiver);
        } else if (hasRole(NATIVE_TOKEN_ROLE, asset)) {
            _receiveNative(amount, receiver);
        } else {
            revert("Attempting to withdraw unknown asset.");
        }
    }

    function _receiveTokens(address asset, uint256 amount, address receiver) private {
        SafeERC20.safeTransfer(IERC20(asset), receiver, amount);
    }

    function _receiveNative(uint256 amount, address receiver) private {
        (bool sent, ) = receiver.call{value: amount}("");
        require(sent, "Failed to send Ether");
    }
}