import * as assert from 'assert';
import * as docpb from "../../../gen/foyle/v1alpha1/doc_pb";
import * as metadata from "../../metadata";
suite('Metadata Test Suite', () => {
    test('test getCellMetadata', () => {
        class TestCase {
            block: docpb.Block;
            expected: metadata.CellMetadata;

            constructor(block: docpb.Block, expected: metadata.CellMetadata) {
                this.block = block;
                this.expected = expected;
            }
        }

        // No traceids
        const block1 = new docpb.Block();
        
        const meta1 = {            
            "traceIds": [],
        };

        // Some traces
        const block2 = new docpb.Block();
        block2.traceIds = ["a", "b"];      
        const meta2 = {            
            "traceIds": ["a", "b"],
        };

        const testCases: TestCase[] = [
            new TestCase(block1, meta1),
            new TestCase(block2, meta2),
        ];

        for (const testCase of testCases) {
            const actual = metadata.getCellMetadata(testCase.block);
            assert.deepStrictEqual(actual, testCase.expected);
        }
    });

    test('test setBlockFromMeta', () => {
        class TestCase {
            meta: metadata.CellMetadata;
            expected: docpb.Block;

            constructor(meta: metadata.CellMetadata, block: docpb.Block) {
                this.meta = meta;
                this.expected = block;
            }
        }

        // No traceids
        const meta1 = {};
        const block1 = new docpb.Block();

        // With traceids
        const meta2 = {
          "traceIds": ["a", "b"],
        };
        const block2 = new docpb.Block();
        block2.traceIds = ["a", "b"];
        

        const testCases: TestCase[] = [
            new TestCase(meta1, block1),
            new TestCase(meta2, block2),
        ];

        for (const testCase of testCases) {
            const actual = new docpb.Block();
            metadata.setBlockFromMeta(actual, testCase.meta);
            assert.deepStrictEqual(actual, testCase.expected);
        }
    });   
});
