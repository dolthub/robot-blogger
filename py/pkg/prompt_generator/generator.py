import uuid
from openai import OpenAI

client = OpenAI()
import mysql.connector
from pkg.config.config import DB_CONFIG, OPENAI_API_KEY
from pkg.prompt_generator.prompts import generate_reverse_prompt

# Ensure this matches the table name in Dolt
TABLE_NAME = "prompts"

def create_prompts_table():
    """Creates the prompts table if it does not exist."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    create_table_query = f"""
    CREATE TABLE IF NOT EXISTS {TABLE_NAME} (
        id VARCHAR(36) NOT NULL PRIMARY KEY,
        blog_md5_hash VARCHAR(36) NOT NULL,
        generated_prompt TEXT NOT NULL,
        model_name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (blog_md5_hash) REFERENCES human_blogs(md5_hash) ON DELETE CASCADE
    );
    """

    cursor.execute(create_table_query)
    connection.commit()
    cursor.close()
    connection.close()
    print("✅ Ensured prompts table exists.")

def fetch_all_blog_md5_hashes():
    """Fetch only blog md5 hashes from the human_blogs table."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    cursor.execute("SELECT md5_hash FROM human_blogs")
    blog_md5_hashes = [row[0] for row in cursor.fetchall()]  # Store only IDs

    print(f"🚀 Found {len(blog_md5_hashes)} blog md5 hashes.")
    cursor.close()
    connection.close()

    return blog_md5_hashes

def fetch_blog_by_md5_hash(blog_md5_hash):
    """Fetch a single blog by md5 hash."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor(dictionary=True)

    cursor.execute("SELECT md5_hash, file_name, file_content_plain FROM human_blogs WHERE md5_hash = %s", (blog_md5_hash,))
    blog = cursor.fetchone()  # Fetch a single blog

    cursor.close()
    connection.close()

    return blog

def store_generated_prompt(blog_md5_hash, prompt, model_name):
    """Stores the generated prompt in the database with model information."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    prompt_id = str(uuid.uuid4())  # Generate unique ID
    insert_query = f"""
    INSERT INTO {TABLE_NAME} (id, blog_md5_hash, generated_prompt, model_name)
    VALUES (%s, %s, %s, %s)
    """
    cursor.execute(insert_query, (prompt_id, blog_md5_hash, prompt, model_name))

    connection.commit()
    cursor.close()
    connection.close()
    print(f"✅ Stored prompt for blog md5 hash {blog_md5_hash} in database (Model: {model_name}).")

def generate_blog_prompt(blog, model_name):
    """Generates a blog-writing prompt from a human-written blog and stores it in DB."""
    blog_md5_hash = blog["md5_hash"]
    print(f"🚀 Blog md5 hash: {blog_md5_hash}")
    file_name = blog["file_name"].replace(".md", "_prompt.txt")
    content = blog["file_content_plain"]
    print(f"🚀 Content Length: {len(content)}")
    prompt = generate_reverse_prompt(content)

    print(f"🚀 Generating reverse-engineered prompt for: {file_name} using model {model_name}...")

    response = client.chat.completions.create(model=model_name,
    messages=[{"role": "system", "content": "You are an expert in AI prompt engineering."},
              {"role": "user", "content": prompt}],
    temperature=0.7)

    generated_prompt = response.choices[0].message.content

    # Store prompt in the database with model name
    store_generated_prompt(blog_md5_hash, generated_prompt, model_name)

    print(f"✅ Prompt stored for blog md5 hash {blog_md5_hash} (Model: {model_name})")
