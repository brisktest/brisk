const moment = require('moment')
const chalk = require('chalk')
const ProgressBar = require('progress')

const passedFmt = chalk.green
const failedFmt = chalk.red
const pendingFmt = chalk.cyan
const infoFmt = chalk.white

export default class JestProgressBarReporter {
    _globalConfig: any
    _options: any
    _numTotalTestSuites: any
    _bar: any
    constructor(globalConfig: any, options: any) {
        this._globalConfig = globalConfig
        this._options = options
        this._numTotalTestSuites
    }

    onRunStart(test: { numTotalTestSuites: any }) {
        const { numTotalTestSuites } = test

        console.log()
        console.log(infoFmt(`Found ${numTotalTestSuites} test suites`))
        this._numTotalTestSuites = numTotalTestSuites
    }

    onTestStart() {
        if (!this._bar) {
            this._bar = new ProgressBar('[:bar] :current/:total :percent', {
                complete: '.',
                incomplete: ' ',
                total: this._numTotalTestSuites,
            })
        }
    }

    onRunComplete(
        test: any,
        results: {
            numFailedTests: any
            numPassedTests: any
            numPendingTests: any
            testResults: any
            numTotalTests: any
            startTime: any
            snapshot: any
        }
    ) {
        const {
            numFailedTests,
            numPassedTests,
            numPendingTests,
            testResults,
            numTotalTests,
            startTime,
            snapshot,
        } = results

        testResults.map(({ failureMessage }: any) => {
            if (failureMessage) {
                console.log(failureMessage)
            }
        })
        console.log(infoFmt(`Ran ${numTotalTests} tests in ${testDuration()}`))
        if (snapshot.failure) {
            console.log(
                `\n${failedFmt(
                    `Obsolete snapshot(s)`
                )} found, run with 'npm test -- -u' to remove them\n`
            )
        }
        if (numPassedTests) {
            console.log(
                this._getStatus('passed') +
                    passedFmt(` ${numPassedTests} passing`)
            )
        }
        if (numFailedTests) {
            console.log(
                this._getStatus('failed') +
                    failedFmt(` ${numFailedTests} failing`)
            )
        }
        if (numPendingTests) {
            console.log(
                this._getStatus('pending') +
                    pendingFmt(` ${numPendingTests} pending`)
            )
        }

        function testDuration() {
            //@ts-ignore
            const delta = moment.duration(moment() - new Date(startTime))
            const seconds = delta.seconds()
            const millis = delta.milliseconds()
            return `${seconds}.${millis} s`
        }
    }

    onTestResult() {
        this._bar.tick()
    }

    _getStatus(status: string) {
        switch (status) {
            case 'passed':
                return passedFmt('✔')
            default:
            case 'failed':
                return failedFmt('✘')
            case 'pending':
                return pendingFmt('-')
        }
    }
}
