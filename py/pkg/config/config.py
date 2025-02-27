import os
from dotenv import load_dotenv

# Load environment variables from .env file (optional)
load_dotenv()

# Database Config
DB_CONFIG = {
    "host": os.getenv("DB_HOST", "0.0.0.0"),
    "user": os.getenv("DB_USER", "root"),
    "password": os.getenv("DB_PASSWORD", ""),
    "database": os.getenv("DB_NAME", "robot_blogger_benchmark_v1"),
    "port": os.getenv("DB_PORT", 3307),
}

# OpenAI API Key
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
