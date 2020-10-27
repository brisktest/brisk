export function prettyPrintOutput(jsonArray: string[]) {
    console.log(JSON.parse(jsonArray.join(' ')))
}
