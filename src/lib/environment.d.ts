import { UrlObject } from 'url'

declare global {
    namespace ChildProcess {
        interface SpawnOptions {
            DOCKER_HOST: UrlObject
            CI_NODE_INDEX: number
            CI_NODE_TOTAL: number
        }
    }
}

// If this file has no import/export statements (i.e. is a script)
// convert it into a module by adding an empty export statement.
export {}
