import os
import json
import uuid
import re
import hashlib
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
            id VARCHAR(36) NOT NULL PRIMARY KEY,
            file_name VARCHAR(2048) NOT NULL UNIQUE,
            metadata JSON,
            file_content_markdown TEXT NOT NULL,
            file_content_plain TEXT NOT NULL,
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
        match = re.match(r"^---\n(.*?)\n---\n", content_markdown, re.DOTALL)
        if match:
            front_matter = match.group(1)

            # Extract "tags" field if present
            tags_match = re.search(r'^"tags":\s*(\[.*?\])', front_matter, re.MULTILINE)
            if tags_match:
                try:
                    tags = json.loads(tags_match.group(1))  # Parse tags as JSON array
                    metadata["tags"] = tags
                except json.JSONDecodeError:
                    print("⚠️ Warning: Failed to parse tags as JSON.")

        return metadata

    def insert_into_db(self, file_path):
        """Reads a file and inserts its contents into the MySQL table."""
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                content_markdown = f.read()

            # Compute MD5 hash of markdown content
            md5_hash = self.compute_md5(content_markdown)

            # Convert Markdown to plaintext
            content_plain = self.markdown_to_plaintext(content_markdown)
            word_count = len(content_plain.split())

            # Generate a unique ID
            file_id = str(uuid.uuid4())

            # Base metadata
            metadata = {
                "file_name": Path(file_path).name,  # Just the base file name
                "size_bytes": os.path.getsize(file_path),
                "created_at": os.path.getctime(file_path),
                "updated_at": os.path.getmtime(file_path),
                "md5_hash": md5_hash  # ✅ Store MD5 hash in metadata
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
            INSERT INTO human_blogs (id, file_name, metadata, file_content_markdown, file_content_plain, word_count)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON DUPLICATE KEY UPDATE
                metadata = VALUES(metadata),
                file_content_markdown = VALUES(file_content_markdown),
                file_content_plain = VALUES(file_content_plain),
                word_count = VALUES(word_count);
            """

            cursor.execute(insert_query, (file_id, Path(file_path).name, metadata_json, content_markdown, content_plain, word_count))
            
            connection.commit()
            cursor.close()
            connection.close()

            print(f"✅ Inserted {file_path} into database.")

        except Exception as e:
            print(f"❌ Error processing {file_path}: {e}")

    def run(self):
        """Main function to process files and insert into MySQL."""
        files = self.get_sorted_files()
        print(f"🔍 Found {len(files)} files to process...")

        for file_path in files:
            self.insert_into_db(file_path)

        print("✅ All files processed successfully.")
