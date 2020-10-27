"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.runMoreFaster = void 0;
const child_process_1 = require("child_process");
const path_1 = __importDefault(require("path"));
const config_1 = require("./config");
const findProjectRoot_1 = require("./findProjectRoot");
const paritionFiles_1 = require("./lib/paritionFiles");
const setupRemote_1 = require("./lib/setupRemote");
exports.runMoreFaster = () => {
    // const chalk = require('chalk')
    // const clear = require('clear')
    // const figlet = require('figlet')
    // const path = require('path')
    // const program = require('commander')
    // //the places we might send jobs
    // const machines = ['127.0.0.1', '192.168.1.2']
    // const priorities = [3, 7, 5] // what way should we split the jobs
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
    // rsync -av  ~/Programming/snackpass-server  ubuntu@34.230.36.44:/workspace
    //nodemon --watch ./ --ext ts,js,json --exec rsync -av  ./  ubuntu@3.93.181.173:/workspace
    //run with env variables
    //docker-compose run -w /workspace -e NPM_TOKEN -e MONGOMS_SYSTEM_BINARY=/usr/bin/mongod -e CLOUDAMQP_URL='rabbitmq:5672' -e REDIS_URL='redis' web npm test
    //  ssh ubuntu@34.230.36.44  sudo /etc/init.d/docker start
    //ssh ubuntu@34.230.36.44 sudo usermod -aG docker ${USER}
    //DOCKER_HOST="ssh://ubuntu@34.230.36.44" docker-compose up -d
    // problems here with mongo on the docker host
    //DOCKER_HOST="ssh://ubuntu@34.230.36.44" docker-compose exec -e REDIS_URL="redis://redis:6379" -e CLOUDAMQP_URL="rabbitmq:5672" -e  NPM_TOKEN=$NPM_TOKEN -e MONGOMS_SYSTEM_BINARY="/usr/bin/mongod" -w /workspace/snackpass-server  web npm run test
    //rsync -av  ~/Programming/snackpass-server  ubuntu@100.25.144.104:/workspace
    //  DOCKER_HOST="ssh://ubuntu@100.25.144.104" docker-compose exec  -e NPM_TOKEN=$NPM_TOKEN  -w /workspace web npm run test
    function sendJobToMachine(job, machine) {
        machine.sendJob(job);
    }
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
    // Connect to n servers - pass an env to each telling them they at n from N
    // Combine the results for local
    // need a config file that specifies the setup
    // or a login for a service which has the setup already and can auto assign N servers and run extremely fast.
    // run a test and get the json in output.file
    //docker-compose exec  -T -w /workspace -e NPM_TOKEN  -e CLOUDAMQP_URL='amqp://rabbitmq:5672' -e REDIS_URL='redis://redis:6379' -e NODE_PATH=./ web node  --max_old_space_size=12240 ./node_modules/jest/bin/jest.js  tests/e2e/scheduledPurchases.test.ts     --config tests/config/jestConfig.js --json  | tee output.file
    // run two and combine the result
    // test_command
    //makes sure everything is working
    //start a rsync background daemon that makes sure everything is up to date. 
    // can run nodemon or fswatch or something - config for what files to watch but I feel like
    // it should be all, although we may want to ignore node_modules for compatibility reasons
    const workDir = process.cwd(); // '/Users/sean/Programming/snackpass-server/.devcontainer/'
    function args(files = '') {
        return [...[
                'exec',
                '-T',
                '-w /workspace',
                '-e TERM=xterm-256color',
            ], ...(config_1.config.remoteEnv || []).map((env) => `-e ${env}`), ...[config_1.config.dockerContainer], ...config_1.config.cmdArgs, ...files];
    }
    const pathToDockerFile = config_1.config.pathToDockerFile || './';
    const rootOfProject = findProjectRoot_1.findProjectRoot(workDir);
    const dockerFilePath = path_1.default.join(rootOfProject, pathToDockerFile);
    console.log(Object.keys(config_1.config.hosts), dockerFilePath);
    setupRemote_1.setupRemote(dockerFilePath);
    const options = {
        env: { ...process.env },
        shell: true,
        cwd: dockerFilePath,
    };
    const processes = Object.keys(config_1.config.hosts).map((host, index) => {
        const hostSpecificOverrides = {
            env: {
                DOCKER_HOST: host.toString(),
                CI_NODE_TOTAL: Object.keys(config_1.config.hosts).length.toString(),
                CI_NODE_INDEX: index.toString(),
                PATH: process.env.PATH,
                HOME: process.env.HOME,
            },
        };
        let files = paritionFiles_1.splitFiles(Object.keys(config_1.config.hosts).length, index, workDir);
        console.log('the files are ', files);
        const proc = child_process_1.spawn(config_1.config.dockerCommand, args(files), {
            ...options,
            ...hostSpecificOverrides,
        });
        return proc;
    });
    const outputFiles = processes.map((p) => {
        const jsonOutputStore = [""];
        configureOutput(p, processes, jsonOutputStore);
        return jsonOutputStore;
    });
    // const output = spawn('docker-compose', args, options)
    // const output2 = spawn('docker-compose', args, options)
    // let testCommands = [output, output2]
    // testCommands.map((c) => configureOutput(c))
    function configureOutput(spawnProcess, processes, jsonOutput) {
        spawnProcess.stdout.on('data', (data) => {
            console.log(`${spawnProcess.pid} stdout: ${data}`);
            jsonOutput.push(data);
        });
        spawnProcess.stderr.on('data', (data) => {
            console.log(`${spawnProcess.pid} stderr: ${data}`);
        });
        spawnProcess.on('error', (error) => {
            console.log(`${spawnProcess.pid} error: ${error.message}`);
        });
        spawnProcess.on('close', (code) => {
            console.log(`${spawnProcess.pid} child process exited with code ${code}`);
            checkFinished(processes);
        });
    }
    //check exit code and exit
    function checkFinished(testCommands) {
        if (testCommands
            .map((subProcess) => {
            return subProcess.exitCode;
        })
            .filter((exitCode) => {
            return exitCode === null;
        }).length === 0) {
            console.log('Seems like all done');
            console.log('exit codes are ', testCommands.map((subProcess) => {
                return subProcess.exitCode;
            }));
            //outputFiles.map((file: string[]) => {prettyPrintOutput(file)})
            process.exit(0);
        }
        else {
            console.log('not done yet');
        }
    }
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
};
