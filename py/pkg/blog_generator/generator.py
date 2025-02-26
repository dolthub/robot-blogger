import os
import uuid
import hashlib
from openai import OpenAI

client = OpenAI()
import mysql.connector
import markdown
from bs4 import BeautifulSoup
from pkg.config.config import DB_CONFIG, OPENAI_API_KEY

# Ensure this matches the table name in Dolt
TABLE_NAME = "generated_blogs"


def create_generated_blogs_table():
    """Creates the generated_blogs table if it does not exist."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    create_table_query = f"""
    CREATE TABLE IF NOT EXISTS {TABLE_NAME} (
        id VARCHAR(36) NOT NULL PRIMARY KEY,
        prompt_id VARCHAR(36) NOT NULL,
        file_content_markdown TEXT NOT NULL,
        file_content_plain TEXT NOT NULL,
        model_name VARCHAR(255) NOT NULL,
        md5_hash VARCHAR(32) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
    );
    """

    cursor.execute(create_table_query)
    connection.commit()
    cursor.close()
    connection.close()
    print("✅ Ensured generated_blogs table exists.")


def fetch_all_prompt_ids():
    """Fetch only prompt IDs from the database."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    cursor.execute("SELECT id FROM prompts")
    prompt_ids = [row[0] for row in cursor.fetchall()]  # Store only IDs

    cursor.close()
    connection.close()

    return prompt_ids


def fetch_prompt_by_id(prompt_id):
    """Fetch a single prompt by ID."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor(dictionary=True)

    cursor.execute(
        "SELECT id, generated_prompt FROM prompts WHERE id = %s", (prompt_id,)
    )
    prompt = cursor.fetchone()  # Fetch a single prompt

    cursor.close()
    connection.close()

    return prompt


def generate_md5_hash(content):
    """Generate an MD5 hash for a given content string."""
    return hashlib.md5(content.encode("utf-8")).hexdigest()


def markdown_to_plaintext(markdown_content):
    """Convert Markdown content to plaintext using BeautifulSoup."""
    html_content = markdown.markdown(markdown_content)  # Convert Markdown to HTML
    soup = BeautifulSoup(html_content, "html.parser")
    return soup.get_text(separator="\n")  # Extract plain text with line breaks


def generate_blog_content(prompt_text, model_name="gpt-4"):
    """Generates blog content in markdown format using OpenAI with streaming."""
    print(f"🚀 Generating blog using model {model_name}...")

    # Ensure prompt_text is a string
    if not isinstance(prompt_text, str):
        print(f"⚠️ Warning: prompt_text is not a string. Converting to string. Type: {type(prompt_text)}")
        prompt_text = str(prompt_text)

    # Call OpenAI API with streaming enabled
    response = client.chat.completions.create(
        model=model_name,
        messages=[
            {
                "role": "system",
                "content": "You are an AI blog writer. Generate a markdown-formatted blog post.",
            },
            {"role": "user", "content": prompt_text},
        ],
        temperature=0.7,
        stream=True  # Enable streaming
    )

    # Initialize empty strings to store the response
    markdown_content = ""
    
    print("📡 Streaming response from LLM...")
    
    for chunk in response:
        if chunk.choices:
            delta = chunk.choices[0].delta
            if delta.content:
                markdown_content += delta.content
                print(delta.content, end="", flush=True)  # Stream to console in real-time
    
    print("\n✅ Finished streaming response.")

    # Convert markdown to plaintext
    plain_text_content = markdown_to_plaintext(markdown_content)

    # Generate MD5 hash
    md5_hash = generate_md5_hash(markdown_content)

    return markdown_content, plain_text_content, md5_hash



def store_generated_blog(
    prompt_id, markdown_content, plain_text_content, model_name, md5_hash
):
    """Stores the generated blog in the database."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    blog_id = str(uuid.uuid4())  # Generate unique ID
    insert_query = f"""
    INSERT INTO {TABLE_NAME} (id, prompt_id, file_content_markdown, file_content_plain, model_name, md5_hash)
    VALUES (%s, %s, %s, %s, %s, %s)
    """
    cursor.execute(
        insert_query,
        (
            blog_id,
            prompt_id,
            markdown_content,
            plain_text_content,
            model_name,
            md5_hash,
        ),
    )

    connection.commit()
    cursor.close()
    connection.close()
    print(
        f"✅ Stored generated blog for prompt ID {prompt_id} in database (Model: {model_name})."
    )
