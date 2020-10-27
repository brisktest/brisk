"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupRemote = exports.checkStatus = void 0;
const child_process_1 = require("child_process");
const config_1 = require("../config");
function startAll(hosts) {
    const procs = hosts.map((host) => {
        child_process_1.execSync(`DOCKER_HOST=${host} docker-compose up -d`);
    });
}
function startHost(host, cwd) {
    child_process_1.execSync(` DOCKER_HOST=${host} docker-compose up -d`, { cwd });
}
function checkStatus(host, dockerFilePath) {
    return (systemSync(`DOCKER_HOST=${host} docker-compose ps -q ${config_1.config.dockerContainer}`, dockerFilePath) ||
        systemSync(`DOCKER_HOST=${host} docker ps -q --no-trunc | grep $(docker-compose ps -q ${config_1.config.dockerContainer})`, dockerFilePath));
}
exports.checkStatus = checkStatus;
function setupRemote(dockerFilePath) {
    Object.keys(config_1.config.hosts).map((host) => {
        // if (checkStatus(host, dockerFilePath)) {
        //     console.log('Everything running')
        //     return true
        // } else {
        //     console.log('need to start things')
        //   addToKnownHost(host)
        return startHost(host, dockerFilePath);
        // }
    });
}
exports.setupRemote = setupRemote;
function addToKnownHost(host) {
    const hostname = new URL(host).hostname;
    child_process_1.execSync(`ssh-keyscan -H ${hostname} >> ~/.ssh/known_hosts`);
}
function systemSync(cmd, cwd) {
    try {
        child_process_1.execSync(cmd, { cwd }).toString();
        return true;
    }
    catch (error) {
        error.status; // Might be 127 in your example.
        error.message; // Holds the message you typically want.
        error.stderr; // Holds the stderr output. Use `.toString()`.
        error.stdout; // Holds the stdout output. Use `.toString()`.
        console.log('non zero exit code from systemSync with command: ', cmd);
        return false;
    }
}
