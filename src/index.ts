import { ChildProcessWithoutNullStreams, spawn } from 'child_process'
import path from 'path'
import { config } from './config'
import { findProjectRoot } from './findProjectRoot'
import { splitFiles } from './lib/paritionFiles'
import { synSetupRemote } from './lib/setupRemote'

export const runMoreFaster = async () => {
    // const chalk = require('chalk')
    // const clear = require('clear')
    // const figlet = require('figlet')
    // const path = require('path')
    // const program = require('commander')

    // clear()
    // console.log(
    //     chalk.keyword('pink')(
    //         figlet.textSync('More Faster', { horizontalLayout: 'full' })
    //     )
    // )

    //start up with a list of things to do (in this case jest tests)
    // split them up by machine
    // send the jobs to the machines
    // wait for feedback
    // if no feedback send to other machines
    // if feedback wait
    // could send jobs in different orders so we finish fast
    // need to rsync between machines
    //
    // machine gets the jobs and starts executing - funneling the results back to the caller who then displays

    //we do have an env setup stage. Init the env. Run npm i. Make sure we have the prequisites - how we specify them is an open question
    // there is a bonjour implementation for finding peers
    // could do a peer to peer thing ?

    // so basic start is - init env - run a test - report back
    // init env
    // we'll need certain things installed.
    // if it's the same build machines - we can just run npm i (if there are changes) and then run the test.
    // but that probably won't be the case. - Docker to have an env ? Heroku buildpacks ? - both ?
    // we will probably need to define the setup somehow
    // could lean on heroku CI or circle.yml
    // Docker container -

    // sudo apt-get update
    // sudo apt-get install \
    //apt-transport-https \
    //ca-certificates \
    //curl \
    //gnupg-agent \
    //software-properties-common
    //sudo add-apt-repository \
    // "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
    // $(lsb_release -cs) \
    // stable"

    //sudo apt-get update
    //sudo apt-get install docker-ce docker-ce-cli containerd.io

    // from https://docs.docker.com/engine/install/ubuntu/

    //ssh ubuntu@34.230.36.44 sudo mkdir -p /workspace
    //ssh ubuntu@34.230.36.44 sudo chown ubuntu /workspace

    // program
    //     .version('0.0.1')
    //     .description("An example CLI for ordering pizza's")
    //     .option('-p, --peppers', 'Add peppers')
    //     .option('-P, --pineapple', 'Add pineapple')
    //     .option('-b, --bbq', 'Add bbq sauce')
    //     .option('-c, --cheese <type>', 'Add the specified type of cheese [marble]')
    //     .option('-C, --no-cheese', 'You do not want any cheese')
    //     .parse(process.argv)

    // console.log('you ordered a pizza with:')
    // if (program.peppers) console.log('  - peppers')
    // if (program.pineapple) console.log('  - pineapple')
    // if (program.bbq) console.log('  - bbq')
    // const cheese: string =
    //     true === program.cheese ? 'marble' : program.cheese || 'no'
    // console.log('  - %s cheese', cheese)

    // if (!process.argv.slice(2).length) {
    //     program.outputHelp()
    // }

    var myArgs = process.argv.slice(2);

    const workDir = process.cwd() // '/Users/sean/Programming/snackpass-server/.devcontainer/'

    const pathToDockerFile = config.pathToDockerFile || './'

    const rootOfProject = findProjectRoot(workDir)
    const dockerFilePath = path.join(rootOfProject, pathToDockerFile)

    console.log(Object.keys(config.hosts), dockerFilePath)

    //setupRemote(dockerFilePath).then(() => {
    synSetupRemote(dockerFilePath).then(() => {
        console.log('hosts setup....')
        const options: { [key: string]: any } = {
            env: { ...process.env },
            shell: true,
            cwd: dockerFilePath,
        }

        const processPromise = Object.keys(config.hosts).map(
            (host: string, index: number) => {
                const hostSpecificOverrides: { env: NodeJS.ProcessEnv } = {
                    env: {
                        DOCKER_HOST: host.toString(),
                        CI_NODE_TOTAL: Object.keys(
                            config.hosts
                        ).length.toString(),
                        CI_NODE_INDEX: index.toString(),
                        PATH: process.env.PATH,
                        HOME: process.env.HOME,
                    },
                }
                // should make this configurable with a test dir maybe?
                // or potentially another test pattern e.g. .test
                // take test dir from file and also file type
                const myPattern = myArgs[0] ? `tests/**/*${myArgs[0]}*.ts` : undefined
                    
                return splitFiles(
                    Object.keys(config.hosts).length,
                    index,
                    workDir,
                    myPattern
                ).then((files) => {
                    console.log('the files are ', files)

                    const proc = spawn(config.dockerCommand, args(files), {
                        ...options,
                        ...hostSpecificOverrides,
                    })
                    return proc
                })
            }
        )

        Promise.all(processPromise).then((processes) => {
            processes.map((p: any) => {
                const jsonOutputStore: string[] = ['']
                configureOutput(p, processes, jsonOutputStore)
                return jsonOutputStore
            })
        })
    })

    // const output = spawn('docker-compose', args, options)
    // const output2 = spawn('docker-compose', args, options)

    // let testCommands = [output, output2]
    // testCommands.map((c) => configureOutput(c))

    //check exit code and exit

    // async function runCommmands(commands:[string]) {
    //     commands.map(async command => {
    //  exec(testCommand,{cwd: workDir, env: {
    //     ...process.env}},  (error: { message: any }, stdout: any, stderr: any) => {
    //     console.log(".")
    //     if (error) {
    //         console.log(`error: ${error.message}`);
    //         return;
    //     }
    //     if (stderr) {
    //         console.log(`stderr: ${stderr}`);
    //         return;
    //     }
    //     console.log(`stdout: ${stdout}`);
    //     return stdout
    // })
    // }
    // )
    // }
    //console.log('running ', testCommand)
    //console.log( runCommmands([testCommand]))
}

function args(files: string[] = []): string[] {
    return [
        ...[
            'exec',
            '-T',
            '-w /workspace',
            '-e TERM=xterm-256color',
           // '--rebuild'

            //"npm run test"
        ],
        ...(config.remoteEnv || []).map((env: string) => `-e ${env}`),
        ...[config.dockerContainer],
        ...config.cmdArgs,
        ...files,
    ]
}
function checkFinished(testCommands: any[]) {
    if (
        testCommands
            .map((subProcess: { exitCode: any }) => {
                return subProcess.exitCode
            })
            .filter((exitCode: null) => {
                return exitCode === null
            }).length === 0
    ) {
        console.log('Seems like all done')
        console.log(
            'exit codes are ',
            testCommands.map((subProcess: { exitCode: any }) => {
                return subProcess.exitCode
            })
        )

        //outputFiles.map((file: string[]) => {prettyPrintOutput(file)})
        process.exit(0)
    } else {
        console.log('not done yet')
    }
}

function configureOutput(
    spawnProcess: ChildProcessWithoutNullStreams,
    processes: ChildProcessWithoutNullStreams[],
    jsonOutput: string[]
) {
    spawnProcess.stdout.on('data', (data: any) => {
       // console.log(`${spawnProcess.pid} stdout: ${data}`)
       //console.log(data.toString())
//jsonOutput.push(data)
    })

    spawnProcess.stderr.on('data', (data: any) => {
         if(process.env.VERBOSE_OUTPUT){
        //console.log(`${spawnProcess.pid} stderr: ${data}`)
        process.stdout.write(data)
         }
    })

    spawnProcess.on('error', (error: { message: any }) => {
        console.log(`${spawnProcess.pid} error: ${error.message}`)
    })

    spawnProcess.on('close', (code: any) => {
        console.log(
            `${spawnProcess.pid} child process exited with code ${code}`
        )
        checkFinished(processes)
    })
}
