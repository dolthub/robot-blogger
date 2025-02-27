def generate_reverse_prompt(content):
    """Constructs a prompt asking the model to generate a blog-writing prompt based on an existing blog."""
    return f"""
    You are an expert in prompt engineering and AI-generated content. Your task is to read a human-written blog post 
    and reverse engineer a high-quality prompt that could have been used to generate that blog post using an AI model.

    **Instructions:**
    - Carefully analyze the blog title and content.
    - Determine the key themes, tone, style, and structure of the blog.
    - Construct a detailed and effective prompt that, when given to an AI model, would result in generating a similar blog.
    - The prompt should be structured in a way that guides an LLM (like GPT-4) to create high-quality, engaging, and informative content.

    Blog Content:
    {content}

    **Example Output (Generated Prompt):**
    "Write a detailed blog post about [TOPIC] with a professional yet engaging tone. The blog should include an introduction, 
    main content with supporting arguments, and a conclusion. Use clear, concise language with examples where necessary."

    Now, based on the input above, generate the best possible blog-writing prompt that could have been used to produce this blog post.
    """
