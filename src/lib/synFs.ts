import chokidar from 'chokidar'
import Rsync from 'rsync'
import url from 'url'
import { config } from '../config'
export function startSyncForHosts(
    hosts: string[],
    localPath: string,
    remotePath: string
) {
    hosts.map((host: string) => {
        runSyncFs(host, localPath, remotePath)
    })
}

export function runSyncFs(
    hostname: string,
    localPath: string,
    remotePath: string
) {
    const justHostname = url.parse(hostname).hostname || ''
    const justUsername = url.parse(hostname).auth
    localPath = localPath + '/'
    //run once at the begininng then watch
    syncFS(justUsername + '@' + justHostname, localPath, remotePath)

    chokidar.watch(localPath).on('change', (path: string) => {
        console.log(path)
        syncFS(justUsername + '@' + justHostname, path.toString(), remotePath + path.toString())
    })
}

// watches for local FS changes and rsync to the remote
function syncFS(hostname: string, cwd: string, workspace: string) {
    console.log(hostname)

    // Build the command
    var rsync = new Rsync()
        .shell('ssh')
        .flags('avz')
        .source(cwd)
        .delete()
        .exclude('.git')
        .destination(`${hostname}:${workspace}`) ;
        
        (config.syncExcludePath || []).forEach((element:string) => {
            rsync.exclude(element)
        });

    const printToStdout = (data: string) => {
        console.log(data)
    }
    const printToStdErr = (data: string) => {
        console.error(data)
    }
    rsync.execute(
        function (error: any, code: any, cmd: any) {
            console.log('command ', cmd, ' exited with code ', code)
            if (error) console.log(error)
        },
        function (data: Buffer) {
            if (data) printToStdout(data.toString())
        },
        function (data: Buffer) {
            if (data) printToStdErr(data.toString())
        }
    )
}
