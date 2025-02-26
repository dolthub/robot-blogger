import os
import openai
import mysql.connector
from pkg.config.config import DB_CONFIG, OPENAI_API_KEY
from pkg.prompt_generator.prompts import generate_prompt

# Directory to store generated blogs
OUTPUT_DIR = "generated_blogs"

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

def generate_prompt(blog):
    """Generates blog prompt using OpenAI and saves it to a file."""
    # blog_id = blog["id"]
    file_name = blog["file_name"].replace(".md", "_ai.md")  # Change file name
    content = blog["file_content_plain"]
    prompt = generate_prompt(file_name, content)
    
    print(f"🚀 Generating blog for: {file_name}...")
    
    response = openai.ChatCompletion.create(
        model="gpt-4",
        messages=[{"role": "system", "content": "You are an expert in prompt engineering and AI-generated content."},
                  {"role": "user", "content": prompt}],
        temperature=0.7
    )
    
    generated_content = response["choices"][0]["message"]["content"]

    # Ensure output directory exists
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    output_path = os.path.join(OUTPUT_DIR, file_name)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(generated_content)

    print(f"✅ Blog saved: {output_path}")

def main():
    """Fetch all blog IDs and process them one at a time."""
    blog_ids = fetch_all_blog_ids()
    
    if not blog_ids:
        print("❌ No blogs found in the database.")
        return

    for blog_id in blog_ids:
        blog = fetch_blog_by_id(blog_id)  # Fetch a single blog at a time
        if blog:
            generate_prompt(blog)
        else:
            print(f"⚠️ Skipping blog ID {blog_id}, not found.")

if __name__ == "__main__":
    main()
