import { exec, execSync } from 'child_process'
import { config, getHostnames } from '../config'

function startAll(hosts: string[]) {
    const procs = hosts.map((host: string) => {
        execSync(`DOCKER_HOST=${host} docker-compose up -d`)
    })
}

function startHost(host: string, cwd: string) {
    return new Promise(function (resolve, reject) {
        exec(
            ` DOCKER_HOST=${host} docker-compose up -d`,
            { cwd },
            (error, stdout, stderr) => {
                if (error) {
                    console.error(`exec error: ${error}`)
                    reject
                    return
                }
                console.log(`stdout: ${stdout}`)
                console.error(`stderr: ${stderr}`)
                resolve
            }
        )
    })
}

export function checkStatus(host: string, dockerFilePath: string) {
    return (
        systemSync(
            `DOCKER_HOST=${host} docker-compose ps -q ${config.dockerContainer}`,
            dockerFilePath
        ) ||
        systemSync(
            `DOCKER_HOST=${host} docker ps -q --no-trunc | grep $(docker-compose ps -q ${config.dockerContainer})`,
            dockerFilePath
        )
    )
}

export async function setupRemote(dockerFilePath: string) {
    if (!process.env.SETUP_REMOTE) return true

    await Promise.all(
        Object.keys(config.hosts).map((host: string) => {
            // if (checkStatus(host, dockerFilePath)) {
            //     console.log('Everything running')
            //     return true
            // } else {
            //     console.log('need to start things')
            //   addToKnownHost(host)
            return startHost(host, dockerFilePath)
            // }
        })
    )
}

export function addKeysForHosts() {
    getHostnames().map(addToKnownHost)
}

function addToKnownHost(host: string) {
    const hostname: string = new URL(host).hostname
    execSync(`ssh-keyscan -H ${hostname} >> ~/.ssh/known_hosts`)
}

function systemSync(cmd: string, cwd: string) {
    try {
        execSync(cmd, { cwd }).toString()
        return true
    } catch (error) {
        error.status // Might be 127 in your example.
        error.message // Holds the message you typically want.
        error.stderr // Holds the stderr output. Use `.toString()`.
        error.stdout // Holds the stdout output. Use `.toString()`.
        console.log('non zero exit code from systemSync with command: ', cmd)
        return false
    }
}
