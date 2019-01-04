# DocGen

Generates HTML documentation from a Google Sheets.

## How to use

_NB: you will need a Google Sheets API token. Follow [the guide](https://developers.google.com/sheets/api/quickstart/go)._
While following the guide, you will obtain a `credentials.json` file. For this application to run, rename it `client_secret.json`.

```
cp .env.example .env
# edit .env to set your spreadsheet's ID and data range

source .env
go run main.go
```

## What should the spreadsheet look like

For now, these are the rules:

- Range is defined in `.env`, `SPREADSHEET_RANGE`
- It should be 3 columns
- Column 0 is used to specify if the row is a title, its level, or if it should be ignored:
  - `x`: the row is ignored
  - `1`, `2` or `3`: 
    - the row is a `N`-level title (`<h1>Title</h1>` for 1, `<h2>Title</h2>` for 2...)
    - the text in column 1 is used as the title
- Column 1 is the title
    - if column 0 is empty, the title level used is 4 (`<h4>Title</h4>`)
- Colmun 2 is the text
    - the text in column 2 is added as text parsed as follows:
		- text encased in `**` is bold (`<strong>Text</strong>`)
		- line starting with `-` is treated as a list element (`<li>`)
		- line containing ` | ` is treated as a table element. The first line of a table is necessarily headings.
		- as in markdown, a line starting with `#` is considered a title. The number of `#` determines the title level.

Current limitations:

- It only supports titles from levels 1 to 3

## License

MIT
