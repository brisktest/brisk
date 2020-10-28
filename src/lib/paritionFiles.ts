import { execSync } from 'child_process'
import glob from 'glob'

function lineCount(filename: string): number {
    const results = execSync(`wc -c <  ${filename}`)
    return parseInt(results.toString())
}

function splitFilesByCount(
    fileArray: string[],
    lineArray: number[],
    startLineCount: number,
    endLineCount: number
) {
    let startIndex = 0

    while (startLineCount > lineArray[startIndex]) {
        startIndex++
    }

    let endIndex = startIndex

    while (lineArray[endIndex] <= endLineCount) {
        endIndex++
    }

    return fileArray.slice(startIndex, endIndex)
}
/**
 * writes a list of files matching the glob pattern to stdout
 * runs only the subset of files which fall within the job, set
 * in the environment variables.
 *
 * CI_NODE_TOTAL is the number of jobs we will be splitting across, 1-indexed
 * CI_NODE_INDEX is the index of the job (subset of files) we should be running, 0-indexed
 */

/**
 * gets a list of files matching the given glob
 * @returns {string[]}
 */
function getFiles(
    numHosts: number,
    thisHost: number,
    globPattern: string,
    cwd: string
): string[] {
    console.debug('the pattern is ', globPattern)
    console.debug('the CWD is ', cwd)
    const allFiles = glob.sync(globPattern, { cwd })

    const allLines = [...allFiles].map(function (filename) {
        return lineCount(filename)
    })

    const add = (a: number, b: number): number => a + b
    const cumulativeLC = allLines.map(function (lc, index) {
        if (index > 0) {
            return allLines.slice(0, index).reduce(add)
        } else {
            return allLines[0]
        }
    })

    // now we split the lines into the nodes
    const nodeLines = cumulativeLC[cumulativeLC.length - 1]

    const avLines = Math.ceil(nodeLines / numHosts)

    const startLineCount = avLines * thisHost

    const endLineCount = startLineCount + avLines

    return splitFilesByCount(
        allFiles,
        cumulativeLC,
        startLineCount,
        endLineCount
    )
}

export function splitFiles(
    numHosts: number,
    thisHost: number,
    cwd: string,
    globPattern = 'tests/**/*.test.ts',
    splitMethod = 'LINE_COUNT'
): Promise<string[]> {
    return new Promise((resolve, reject) => {
        let files: string[] = []

        switch (splitMethod) {
            case 'LINE_COUNT':
                files = getFiles(numHosts, thisHost, globPattern, cwd)
                break
            default:
                console.error('No splitter provided')
        }

        resolve(files)
    })
}
