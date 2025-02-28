import evaluate
import pandas as pd
import nltk
nltk.download("punkt") 

def load_blogs(file_path):
    """Load blog texts from a CSV file."""
    try:
        df = pd.read_csv(file_path)
        return df["file_content_plain"].tolist()  # Ensure correct column name
    except Exception as e:
        print(f"❌ Error loading {file_path}: {e}")
        exit(1)


def benchmark(human_blogs_path, generated_blogs_path):
    # Load benchmark datasets
    human_blogs = load_blogs(human_blogs_path)
    generated_blogs = load_blogs(generated_blogs_path)

    # Initialize evaluation metrics
    bleu = evaluate.load("bleu")
    rouge = evaluate.load("rouge")
    bertscore = evaluate.load("bertscore")

    # Compute scores
    bleu_score = bleu.compute(predictions=generated_blogs, references=human_blogs)
    rouge_score = rouge.compute(predictions=generated_blogs, references=human_blogs)
    bert_score = bertscore.compute(
        predictions=generated_blogs, references=human_blogs, lang="en"
    )

    # Print evaluation results
    print("BLEU Score:", bleu_score)
    print("ROUGE Score:", rouge_score)
    print("BERTScore:", bert_score["f1"])
