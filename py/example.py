from pkg.file_ingestor.ingestor import FileIngestor
from pkg.file_ingestor.config import DB_CONFIG

# Define a filter function (e.g., only `.md` files)
def filter_markdown_files(filename):
    return filename.endswith(".md")

# Initialize and run the ingestor with "blog_post" type
ingestor = FileIngestor(DB_CONFIG, directory="/Users/dustin/src/ld/web/packages/blog/src/pages", filter_func=filter_markdown_files, doc_type="blog_post")
ingestor.run()
