#!/usr/bin/env python3

import argparse
import sys
from pkg.file_ingestor.ingestor import FileIngestor
from pkg.config.config import DB_CONFIG
from pkg.prompt_generator.generator import fetch_all_blog_ids, fetch_blog_by_id, generate_blog_prompt


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


def run_prompt_generator(limit=None, model=None):
    """Runs the prompt generator to reverse-engineer prompts from human-written blogs."""
    blog_ids = fetch_all_blog_ids()

    if not blog_ids:
        print("❌ No blogs found in the database.")
        return

    if limit:
        blog_ids = blog_ids[:limit]

    for blog_id in blog_ids:
        blog = fetch_blog_by_id(blog_id)  # Fetch a single blog at a time
        if blog:
            generate_blog_prompt(blog, model)
        else:
            print(f"⚠️ Skipping blog ID {blog_id}, not found.")


def main():
    parser = argparse.ArgumentParser(description="Benchmark Tool CLI")

    # Define available subcommands
    parser.add_argument(
        "command",
        choices=["ingest", "generate-prompt"],
        help="Command to execute"
    )

    # Arguments for 'ingest' command
    parser.add_argument(
        "--dir",
        type=str,
        required=False,
        help="Directory to ingest files from (required for ingest)",
    )
    parser.add_argument(
        "--doc-type",
        type=str,
        required=False,
        help="Type of document to process (required for ingest)",
    )

    # Arguments for 'generate-prompt' command
    parser.add_argument(
        "--limit",
        type=int,
        required=False,
        help="Limit the number of blog prompts generated (optional)"
    )

    parser.add_argument(
        "--model",
        type=str,
        required=True,
        help="Model to use for prompt generation (optional)"
    ) 
    args = parser.parse_args()

    if args.command == "ingest":
        if not args.dir or not args.doc_type:
            print("❌ Error: --dir and --doc-type are required for 'ingest'.", file=sys.stderr)
            sys.exit(1)
        run_ingestor(args.dir, args.doc_type)

    elif args.command == "generate-prompt":
        run_prompt_generator(args.limit, args.model)


if __name__ == "__main__":
    main()
