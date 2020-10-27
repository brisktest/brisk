"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.runSyncFs = exports.startSyncForHosts = void 0;
const chokidar_1 = __importDefault(require("chokidar"));
const rsync_1 = __importDefault(require("rsync"));
const url_1 = __importDefault(require("url"));
function startSyncForHosts(hosts, localPath, remotePath) {
    hosts.map((host) => {
        runSyncFs(host, localPath, remotePath);
    });
}
exports.startSyncForHosts = startSyncForHosts;
function runSyncFs(hostname, localPath, remotePath) {
    const justHostname = url_1.default.parse(hostname).hostname || '';
    const justUsername = url_1.default.parse(hostname).auth;
    localPath = localPath + '/';
    syncFS(justUsername + '@' + justHostname, localPath, remotePath);
    chokidar_1.default.watch(localPath).on('change', (event, path) => {
        console.log(event, path);
        syncFS(justUsername + '@' + justHostname, localPath, remotePath);
    });
}
exports.runSyncFs = runSyncFs;
// watches for local FS changes and rsync to the remote
function syncFS(hostname, cwd, workspace) {
    console.log(hostname);
    //const command = `nodemon --watch ./ --ext . --exec rsync -av  ./  ${hostname}:${workspace}`
    //    const proc = spawn(command, { cwd, ...process.env })
    // Build the command
    var rsync = new rsync_1.default()
        .shell('ssh')
        .flags('av')
        .source(cwd)
        .destination(`${hostname}:${workspace}`);
    const finishedCallback = (code, cmd) => {
        console.log('command ', cmd, ' exited with code ', code);
    };
    // Execute the command
    const printToStdout = (data) => {
        console.log(data);
    };
    const printToStdErr = (data) => {
        console.error(data);
    };
    rsync.execute(function (error, code, cmd) {
        console.log('command ', cmd, ' exited with code ', code);
        if (error)
            console.log(error);
    }, function (data) {
        printToStdout(data.toString());
        // do things like parse progress
    }, function (data) {
        printToStdErr(data.toString());
    });
    // proc.on('close', (data: any) => {
    //     console.debug(data)
    // })
    // proc.on('disconnect', printToStdout)
    // proc.on('error', printToStdout)
    // proc.on('message', printToStdout)
    // proc.on('exit', printToStdout)
    // One-liner for current directory
}
