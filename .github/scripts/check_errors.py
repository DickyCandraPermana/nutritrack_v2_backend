import os
import subprocess
import requests
from google import genai # Menggunakan SDK terbaru
from typing import Optional

# Inisialisasi Gemini Client
# Versi terbaru tidak perlu genai.configure, langsung lewat Client
client = genai.Client(api_key=os.getenv("GEMINI_API_KEY"))

def get_git_diff() -> Optional[str]:
    """Mengambil perubahan baris di commit terakhir."""
    try:
        # Mengambil diff antara commit sekarang (HEAD) dan sebelumnya (HEAD^)
        diff = subprocess.check_output(["git", "diff", "HEAD^", "HEAD"]).decode("utf-8")
        return diff if diff.strip() else None
    except Exception as e:
        print(f"Error fetching diff: {e}")
        return None

def analyze_with_gemini(diff_text: str) -> str:
    """Kirim diff ke Gemini untuk analisa error."""
    prompt = f"""
    You are a Senior Software Engineer. Analyze the following Git Diff for:
    1. Logic errors or bugs.
    2. Security vulnerabilities.
    3. Performance issues.

    If you find any issues, provide a concise report with:
    - 🔍 **Issue**: Description of the problem.
    - 💡 **Suggestion**: How to fix it.
    - 🛠️ **Code Snippet**: Example of the fixed code.

    If the code looks good and no issues are found, reply ONLY with the word "CLEAR".

    GIT DIFF:
    {diff_text}
    """

    # Menggunakan model gemini-2.0-flash untuk kecepatan dan stabilitas API
    response = client.models.generate_content(
        model="gemini-2.0-flash", 
        contents=prompt
    )
    return response.text

def create_github_issue(report: str):
    """Membuat issue di GitHub repository."""
    repo = os.getenv("GITHUB_REPOSITORY")
    token = os.getenv("GITHUB_TOKEN")
    url = f"https://api.github.com/repos/{repo}/issues"
    
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json"
    }
    
    data = {
        "title": "🚨 AI Code Review: Potential Issues Detected",
        "body": f"Gemini AI found some potential issues in the latest commit:\n\n{report}",
        "labels": ["bug", "ai-review"]
    }
    
    res = requests.post(url, json=data, headers=headers)
    if res.status_code == 201:
        print("✅ Issue successfully created.")
    else:
        print(f"❌ Failed to create issue: {res.status_code} - {res.text}")

def main():
    print("Checking for changes...")
    diff = get_git_diff()
    
    if not diff:
        print("No significant changes or no previous commit found to compare.")
        return

    print("Analyzing code with Gemini 2.0 Flash...")
    try:
        analysis = analyze_with_gemini(diff)

        # Cek apakah AI memberikan lampu hijau
        if "CLEAR" in analysis.upper() and len(analysis.strip()) < 15:
            print("✨ Everything looks good! No issue created.")
        else:
            print("⚠️ Issues found! Creating GitHub Issue...")
            create_github_issue(analysis)
    except Exception as e:
        print(f"❌ Gemini API Error: {e}")

if __name__ == "__main__":
    main()
