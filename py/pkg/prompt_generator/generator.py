import uuid
import openai
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
        blog_id VARCHAR(36) NOT NULL,
        generated_prompt TEXT NOT NULL,
        model_name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (blog_id) REFERENCES human_blogs(id) ON DELETE CASCADE
    );
    """
    
    cursor.execute(create_table_query)
    connection.commit()
    cursor.close()
    connection.close()
    print("✅ Ensured prompts table exists.")

def fetch_all_blog_ids():
    """Fetch only blog IDs from the human_blogs table."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    cursor.execute("SELECT id FROM human_blogs")
    blog_ids = [row[0] for row in cursor.fetchall()]  # Store only IDs

    cursor.close()
    connection.close()

    return blog_ids

def fetch_blog_by_id(blog_id):
    """Fetch a single blog by ID."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor(dictionary=True)

    cursor.execute("SELECT id, file_name, file_content_plain FROM human_blogs WHERE id = %s", (blog_id,))
    blog = cursor.fetchone()  # Fetch a single blog

    cursor.close()
    connection.close()

    return blog

def store_generated_prompt(blog_id, prompt, model_name):
    """Stores the generated prompt in the database with model information."""
    connection = mysql.connector.connect(**DB_CONFIG)
    cursor = connection.cursor()

    prompt_id = str(uuid.uuid4())  # Generate unique ID
    insert_query = f"""
    INSERT INTO {TABLE_NAME} (id, blog_id, generated_prompt, model_name)
    VALUES (%s, %s, %s, %s)
    """
    cursor.execute(insert_query, (prompt_id, blog_id, prompt, model_name))

    connection.commit()
    cursor.close()
    connection.close()
    print(f"✅ Stored prompt for blog ID {blog_id} in database (Model: {model_name}).")

def generate_blog_prompt(blog, model_name):
    """Generates a blog-writing prompt from a human-written blog and stores it in DB."""
    blog_id = blog["id"]
    file_name = blog["file_name"].replace(".md", "_prompt.txt")
    content = blog["file_content_plain"]

    prompt = generate_reverse_prompt(file_name, content)
    
    print(f"🚀 Generating reverse-engineered prompt for: {file_name} using model {model_name}...")

    response = openai.ChatCompletion.create(
        model=model_name,
        messages=[{"role": "system", "content": "You are an expert in AI prompt engineering."},
                  {"role": "user", "content": prompt}],
        temperature=0.7
    )
    
    generated_prompt = response["choices"][0]["message"]["content"]

    # Store prompt in the database with model name
    store_generated_prompt(blog_id, generated_prompt, model_name)

    print(f"✅ Prompt stored for blog ID {blog_id} (Model: {model_name})")
