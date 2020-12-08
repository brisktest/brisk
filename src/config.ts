import { findProjectRoot } from './findProjectRoot'

let config: { [key: string]: any }
const projectRoot = findProjectRoot(process.cwd.toString())

try {
    config = require(projectRoot + '/brisk.json')
} catch (e) {
    console.log(e.message)
    config = {}
}
config.projectRoot = projectRoot
console.log(config)

export function getHostnames() {
    return Object.keys(config.hosts)
}
export { config }
