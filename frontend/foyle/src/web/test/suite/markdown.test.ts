import * as assert from 'assert';
import * as vscode from 'vscode';
import { parseMarkdown, writeCellsToMarkdown } from '../../markdownParser';
import { rawToNotebookCellData } from '../../markdown';
import { bashLang } from '../../constants';

// Originally copied from 
// https://github.com/microsoft/vscode-markdown-notebook/blob/main/src/test/suite/index.ts

suite('parseMarkdown', () => {
	test('markdown cell', () => {
		const cells = parseMarkdown('# hello');
		assert.strictEqual(cells.length, 1);
		assert.strictEqual(cells[0].content, '# hello');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '');
	});

	test('markdown cell, w/ whitespace', () => {
		const cells = parseMarkdown('\n\n# hello\n');
		assert.strictEqual(cells.length, 1);
		assert.strictEqual(cells[0].content, '# hello');
		assert.strictEqual(cells[0].leadingWhitespace, '\n\n');
		assert.strictEqual(cells[0].trailingWhitespace, '\n');
	});

	test('2 markdown cells', () => {
		const cells = parseMarkdown('# hello\n\n# goodbye\n');
		assert.strictEqual(cells.length, 2);
		assert.strictEqual(cells[0].content, '# hello');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '\n\n');

		assert.strictEqual(cells[1].content, '# goodbye');
		assert.strictEqual(cells[1].leadingWhitespace, '');
		assert.strictEqual(cells[1].trailingWhitespace, '\n');
	});

	test('code cell', () => {
		const cells = parseMarkdown('```js\nlet x = 1;\n```');
		assert.strictEqual(cells.length, 1);
		assert.strictEqual(cells[0].content, 'let x = 1;');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '');
		assert.strictEqual(cells[0].language, 'javascript');
	});

  test('code cell no langid', () => {
		const cells = parseMarkdown('```\necho hello world\n```');
		assert.strictEqual(cells.length, 1);
		assert.strictEqual(cells[0].content, 'echo hello world');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '');
		assert.strictEqual(cells[0].language, bashLang);
	});

	test('code cell, w/whitespace', () => {
		const cells = parseMarkdown('\n\n```js\nlet x = 1;\n```\n\n');
		assert.strictEqual(cells.length, 1);
		assert.strictEqual(cells[0].content, 'let x = 1;');
		assert.strictEqual(cells[0].leadingWhitespace, '\n\n');
		assert.strictEqual(cells[0].trailingWhitespace, '\n\n');
	});

	test('code cell, markdown', () => {
		const cells = parseMarkdown('```js\nlet x = 1;\n```\n\n# hello\nfoo\n');
		assert.strictEqual(cells.length, 2);
		assert.strictEqual(cells[0].content, 'let x = 1;');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '\n\n');

		assert.strictEqual(cells[1].content, '# hello\nfoo');
		assert.strictEqual(cells[1].leadingWhitespace, '');
		assert.strictEqual(cells[1].trailingWhitespace, '\n');
	});

	test('markdown, code cell', () => {
		const cells = parseMarkdown('# hello\nfoo\n\n```js\nlet x = 1;\n```\n');
		assert.strictEqual(cells.length, 2);

		assert.strictEqual(cells[0].content, '# hello\nfoo');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '\n\n');

		assert.strictEqual(cells[1].content, 'let x = 1;');
		assert.strictEqual(cells[1].leadingWhitespace, '');
		assert.strictEqual(cells[1].trailingWhitespace, '\n');
	});

	test('markdown, code cell with no whitespace between', () => {
		const cells = parseMarkdown('# hello\nfoo\n```js\nlet x = 1;\n```\n');
		assert.strictEqual(cells.length, 2);

		assert.strictEqual(cells[0].content, '# hello\nfoo');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '\n');

		assert.strictEqual(cells[1].content, 'let x = 1;');
		assert.strictEqual(cells[1].leadingWhitespace, '');
		assert.strictEqual(cells[1].trailingWhitespace, '\n');
	});

	test('indented code cell', () => {
		const cells = parseMarkdown('    ```js\n    // indented js block\n    ```\n# More content');
		assert.strictEqual(cells.length, 2);

		assert.strictEqual(cells[0].content, '// indented js block');
		assert.strictEqual(cells[0].indentation, '    ');

		assert.strictEqual(cells[1].content, '# More content');
	});

	test('CRLF', () => {
		const cells = parseMarkdown('```js\r\nlet x = 1;\r\n```\r\n\r\n# hello\r\nfoo\r\n');
		assert.strictEqual(cells.length, 2);
		assert.strictEqual(cells[0].content, 'let x = 1;');
		assert.strictEqual(cells[0].leadingWhitespace, '');
		assert.strictEqual(cells[0].trailingWhitespace, '\n\n');

		assert.strictEqual(cells[1].content, '# hello\nfoo');
		assert.strictEqual(cells[1].leadingWhitespace, '');
		assert.strictEqual(cells[1].trailingWhitespace, '\n');
	});

	test('empty', () => {
		const cells = parseMarkdown('');
		assert.strictEqual(cells.length, 0);
	});
});

suite('writeMarkdown', () => {
	function testWriteMarkdown(markdownStr: string) {
		const cells = parseMarkdown(markdownStr)
			.map(rawToNotebookCellData);
		assert.strictEqual(writeCellsToMarkdown(cells), markdownStr);
	}

	suite('writeMarkdown', () => {
		test('idempotent', () => {
			testWriteMarkdown('# hello');
			testWriteMarkdown('\n\n# hello\n\n');
			testWriteMarkdown('# hello\n\ngoodbye');

			testWriteMarkdown('```js\nlet x = 1;\n```\n\n# hello\n');
			testWriteMarkdown('```js\nlet x = 1;\n```\n\n```ts\nlet y = 2;\n```\n# hello\n');

			testWriteMarkdown('    ```js\n    // indented code cell\n    ```');
		});

		test('append markdown cells', () => {
			const cells = parseMarkdown(`# hello`)
				.map(rawToNotebookCellData);
			cells.push(<vscode.NotebookCellData>{
				kind: vscode.NotebookCellKind.Markup,
				languageId: 'markdown',
				metadata: {},
				outputs: [],
				value: 'foo'
			});
			cells.push(<vscode.NotebookCellData>{
				kind: vscode.NotebookCellKind.Markup,
				languageId: 'markdown',
				metadata: {},
				outputs: [],
				value: 'bar'
			});

			assert.strictEqual(writeCellsToMarkdown(cells), `# hello\n\nfoo\n\nbar\n`);
		});

		test('append code cells', () => {
			const cells = parseMarkdown('```ts\nsome code\n```')
				.map(rawToNotebookCellData);
			cells.push(<vscode.NotebookCellData>{
				kind: vscode.NotebookCellKind.Code,
				languageId: 'typescript',
				metadata: {},
				outputs: [],
				value: 'foo'
			});
			cells.push(<vscode.NotebookCellData>{
				kind: vscode.NotebookCellKind.Code,
				languageId: 'typescript',
				metadata: {},
				outputs: [],
				value: 'bar'
			});

			assert.strictEqual(writeCellsToMarkdown(cells), '```ts\nsome code\n```\n\n```ts\nfoo\n```\n\n```ts\nbar\n```\n');
		});

		test('insert cells', () => {
			const cells = parseMarkdown('# Hello\n\n## Header 2')
				.map(rawToNotebookCellData);
			cells.splice(1, 0, <vscode.NotebookCellData>{
				kind: vscode.NotebookCellKind.Code,
				languageId: 'typescript',
				metadata: {},
				outputs: [],
				value: 'foo'
			});
			cells.splice(2, 0, <vscode.NotebookCellData>{
				kind: vscode.NotebookCellKind.Code,
				languageId: 'typescript',
				metadata: {},
				outputs: [],
				value: 'bar'
			});

			assert.strictEqual(writeCellsToMarkdown(cells), '# Hello\n\n```ts\nfoo\n```\n\n```ts\nbar\n```\n\n## Header 2');
		});
	});
});