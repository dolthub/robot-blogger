import os
import json
import uuid
import re
import hashlib
import sys
import mysql.connector
import markdown
from bs4 import BeautifulSoup
from pathlib import Path

class FileIngestor:
    def __init__(self, db_config, directory, filter_func, doc_type=None):
        """
        Initializes the FileIngestor.

        :param db_config: Dictionary with MySQL connection details.
        :param directory: Path to the directory containing files.
        :param filter_func: Function to filter files (should return True/False).
        :param doc_type: Type of document being processed (e.g., "blog_post").
        """
        self.db_config = db_config
        self.directory = directory
        self.filter_func = filter_func
        self.doc_type = doc_type
        self.ensure_table_exists()

    def ensure_table_exists(self):
        """Ensures the expected table schema is created in the MySQL database."""
        create_table_query = """
        CREATE TABLE IF NOT EXISTS human_blogs (
            md5_hash VARCHAR(32) NOT NULL PRIMARY KEY,
            file_name VARCHAR(255) NOT NULL UNIQUE,
            metadata JSON,
            file_content_markdown LONGTEXT NOT NULL,
            file_content_plain LONGTEXT NOT NULL,
            word_count INT NOT NULL
        );
        """

        try:
            connection = mysql.connector.connect(**self.db_config)
            cursor = connection.cursor()
            cursor.execute(create_table_query)
            connection.commit()
            cursor.close()
            connection.close()
            print("✅ Table 'human_blogs' ensured in database.")
        except Exception as e:
            print(f"❌ Error ensuring table exists: {e}")

    def get_sorted_files(self):
        """Collects and sorts files lexicographically based on filter criteria."""
        files = [os.path.join(self.directory, f) for f in os.listdir(self.directory) if self.filter_func(f)]
        return sorted(files)  # Lexicographic sorting

    def markdown_to_plaintext(self, markdown_content):
        """Converts Markdown to plaintext."""
        html = markdown.markdown(markdown_content)
        soup = BeautifulSoup(html, "html.parser")
        return soup.get_text()

    def compute_md5(self, content):
        """Computes the MD5 hash of the given content."""
        return hashlib.md5(content.encode('utf-8')).hexdigest()

    def extract_metadata_from_blog(self, content_markdown):
        """
        Extracts metadata from YAML front matter in Markdown files.

        :param content_markdown: The raw markdown content.
        :return: Dictionary containing extracted metadata (tags, etc.).
        """
        metadata = {}

        # Look for front matter enclosed by "---\n ... ---\n"
        match = re.match(r"^---\s*\n(.*?)\n---\s*\n", content_markdown, re.DOTALL)
        if match:
            front_matter = match.group(1)

            # Extract "tags" field if present
            tags = self.extract_tags(front_matter)
            if tags:
                metadata["tags"] = tags
        return metadata

    def extract_tags(self, front_matter):
        """Extracts the 'tags' field from YAML front matter."""
        tags_match = re.search(r'^tags:\s*(\[.*?\]|".*?"|\S+)', front_matter, re.MULTILINE)
        if tags_match:
            raw_tags = tags_match.group(1).strip()

            # Try parsing JSON list if it matches a JSON format
            if raw_tags.startswith("["):
                try:
                    tags = json.loads(raw_tags)  # Parse JSON array
                except json.JSONDecodeError:
                    print("⚠️ Warning: Failed to parse tags as JSON.")
                    tags = [raw_tags]
            else:
                # Remove quotes if present
                tags = [raw_tags.strip('"')]

            return tags
        return None  # No tags found

    def insert_into_db(self, file_path):
        """Reads a file and inserts its contents into the MySQL table."""
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                content_markdown = f.read()

                # # Fix escape sequences
                # content_markdown = content_markdown.replace("\\", "\\\\")

                # # Replace "mysql>" at the start of a line with "sql prompt:"
                # content_markdown = re.sub(r"^mysql>\s*", "sql prompt: ", content_markdown, flags=re.MULTILINE)

                start = 1388
                end = 1500

                # content_markdown = content_markdown[0:end]



            # Compute MD5 hash of markdown content
            md5_hash = self.compute_md5(content_markdown)

            # Convert Markdown to plaintext
            content_plain = self.markdown_to_plaintext(content_markdown)
            word_count = len(content_plain.split())

            # # Generate a unique ID
            # file_id = str(uuid.uuid4())

            # Base metadata
            metadata = {
                "file_name": Path(file_path).name,  # Just the base file name
                "size_bytes": os.path.getsize(file_path),
                # "md5_hash": md5_hash,  # ✅ Store MD5 hash in metadata
                "doc_type": self.doc_type
            }

            # If the doc type is "blog_post", extract additional metadata
            if self.doc_type == "blog_post":
                extracted_metadata = self.extract_metadata_from_blog(content_markdown)
                metadata.update(extracted_metadata)  # Merge metadata

                # Check if the "generated" tag is present
                if "tags" in metadata and "generated" in metadata["tags"]:
                    print(f"⏭️ Skipping generated blog: {Path(file_path).name}")
                    return  # Skip insertion and continue

            metadata_json = json.dumps(metadata)

            connection = mysql.connector.connect(**self.db_config)

            cursor = connection.cursor()

            insert_query = """
            INSERT INTO human_blogs (md5_hash, file_name, metadata, file_content_markdown, file_content_plain, word_count)
            VALUES (%s, %s, %s, %s, %s, %s);
            """

            # content_markdown = content_markdown[0:start]
            # content_plain = content_plain[0:start]

            cursor.execute(insert_query, (md5_hash, Path(file_path).name, metadata_json, content_markdown, content_plain, word_count))
            
            connection.commit()
            cursor.close()
            connection.close()

            print(f"✅ Inserted {file_path} into database.")

        except Exception as e:
            print(f"❌ Error processing {file_path}: {e}")
            print(f"❌ Content Markdown: {repr(content_markdown)}")
            print()
            print(f"❌ Content Plain: {repr(content_plain)}")
            # print(f"🔍 Problematic snippet [{start}:{end}]:")
            # print(f"Last 6 characters markdown: {content_markdown[-6:start]}")
            # print(f"Last 6 characters plain: {content_plain[-6:start]}")
            self.panic(f"STOPPING: {file_path}: {e}")

    def panic(message):
        print(f"❌ PANIC: {message}", file=sys.stderr)
        sys.exit(1)  # Exit with failure

    def run(self):
        """Main function to process files and insert into MySQL."""
        files = self.get_sorted_files()
        print(f"🔍 Found {len(files)} files to process...")

        for file_path in files:
            self.insert_into_db(file_path)

        print("✅ All files processed successfully.")
