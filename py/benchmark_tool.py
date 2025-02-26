#!/usr/bin/env python3

import argparse
import sys
from pkg.file_ingestor.ingestor import FileIngestor
from pkg.file_ingestor.config import DB_CONFIG


# Define a filter function (e.g., only `.md` files)
def filter_markdown_files(filename):
    return filename.endswith(".md")


def run_ingestor(directory, doc_type):
    """Runs the file ingestor with the specified directory and document type."""
    if not directory.strip():
        print("❌ Error: --dir must not be empty.", file=sys.stderr)
        sys.exit(1)
    if not doc_type.strip():
        print("❌ Error: --doc-type must not be empty.", file=sys.stderr)
        sys.exit(1)

    print(f"🚀 Running ingestor in {directory} for document type: {doc_type}")
    ingestor = FileIngestor(
        DB_CONFIG,
        directory=directory,
        filter_func=filter_markdown_files,
        doc_type=doc_type,
    )
    ingestor.run()


def main():
    parser = argparse.ArgumentParser(description="File Ingestor CLI")

    # Add 'ingest' as a positional command
    parser.add_argument("command", choices=["ingest"], help="Command to execute")

    # Add required arguments for 'ingest'
    parser.add_argument(
        "--dir",
        type=str,
        required=True,
        help="Directory to ingest files from (required)",
    )
    parser.add_argument(
        "--doc-type",
        type=str,
        required=True,
        help="Type of document to process (required)",
    )

    args = parser.parse_args()

    if args.command == "ingest":
        run_ingestor(args.dir, args.doc_type)


if __name__ == "__main__":
    main()
