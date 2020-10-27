// Simple version of `find-project-root`
// https://github.com/kirstein/find-project-root/blob/master/index.js

const fs = require('fs')
const path = require('path')

const MARKERS = ['.git', '.hg']

const markerExists = (directory: any) => {
    return MARKERS.some((mark) => {
        const fileMark = path.join(directory, mark)
        return fs.existsSync(fileMark)
    })
}

export function findProjectRoot(directory: any) {
    while (!markerExists(directory)) {
        const parentDirectory = path.resolve(directory, '..')
        if (parentDirectory === directory) {
            break
        }
        directory = parentDirectory
    }

    return directory
}
