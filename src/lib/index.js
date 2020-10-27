#!/usr/bin/env node
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
const { exec } = require("child_process");
const testCommand = "DOCKER_HOST=ssh://ubuntu@54.83.94.77 docker-compose exec  -T -w /workspace -e NPM_TOKEN  -e CLOUDAMQP_URL='amqp://rabbitmq:5672' -e REDIS_URL='redis://redis:6379' -e NODE_PATH=./ web node  --max_old_space_size=12240 ./node_modules/jest/bin/jest.js  tests/e2e/scheduledPurchases.test.ts     --config tests/config/jestConfig.js --json  ";
const workDir = "/Users/sean/Programming/snackpass-server/.devcontainer";
function runCommmands(commands) {
    commands.map(command => {
        exec(command, { cwd: workDir }, (error, stdout, stderr) => {
            console.log(".");
            if (error) {
                console.log(`error: ${error.message}`);
                return;
            }
            if (stderr) {
                console.log(`stderr: ${stderr}`);
                return;
            }
            console.log(`stdout: ${stdout}`);
            return stdout;
        });
    });
}
console.log("running ", testCommand);
console.log(await runCommmands([testCommand]));
console.log("finished");
export {};
