import { exec } from 'child_process'
import glob from 'glob'
import { config } from '../config'

var shellParser = require('node-shell-parser')
//need to make this promises
function lineCount(filename: string): Promise<number> {
    return new Promise((resolve, reject) => {
        console.log('filename is ', filename)
        exec(`wc -c <  ${filename}`, (error, results, err) => {
            console.log(error)
            console.log(results)
            console.log(err)
            const res = parseInt(results.toString())
            resolve(res)
        })
    })
}

function wordCountForFiles(
    filenames: string[]
): Promise<Record<string, string>> {
    return new Promise((resolve, reject) => {
        
        //console.debug('filesnames are ', filenames)
        exec(`wc -c ${filenames.map((filename) => `"${filename}"` ).join(' ')}`, {}, (error, stdout, stderr) => {
            if (error) {
                console.error(error)
            }
            //console.log('split by lines')
            //console.error(stderr)
            const lines = stdout.split('\n')

            const output = Object.fromEntries(
                lines.map((l) => {
                    return l.split(' ').reverse()
                })
            )

            resolve(output)
        })
    })
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
    //console.log('near end of split')

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
): Promise<string[]> {
    return new Promise(async (resolve, reject) => {
        console.debug('the pattern is ', globPattern)
        console.debug('the CWD is ', cwd)
        console.debug('about to glob')
        console.log('cwd == ', cwd)

        if (config.testPathIgnorePatterns)
            console.debug('Ignore pattern is ', config.testPathIgnorePatterns)
        glob(
            globPattern,
            { cwd, ignore: config.testPathIgnorePatterns || false },
            (error, allFiles) => {
                if (error) {
                    console.error(error)
                }

                //console.log('allfiles are :', allFiles)
                wordCountForFiles(allFiles).then((filesToCount) => {
                    const allLines = [...allFiles].map(function (filename) {
                        return parseInt(filesToCount[filename])
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
                    const splitFiles = splitFilesByCount(
                        allFiles,
                        cumulativeLC,
                        startLineCount,
                        endLineCount
                    )

                    return resolve(splitFiles)
                })
            }
        )
    })
}

export function splitFiles(
    numHosts: number,
    thisHost: number,
    cwd: string,
    globPattern = config.testPattern || 'tests/**/*.test.ts',
    splitMethod = 'LINE_COUNT'
): Promise<string[]> {
    return new Promise((resolve, reject) => {
        let files: string[] = []

        switch (splitMethod) {
            case 'LINE_COUNT':
                getFiles(numHosts, thisHost, globPattern, cwd).then(resolve)
                break
            default:
                console.error('No splitter provided')
        }
    })
}
