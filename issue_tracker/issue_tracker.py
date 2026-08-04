import json
import os
from datetime import datetime

DB_FILE = os.path.join(os.path.dirname(__file__), "issues.json")

def load_issues():
    if not os.path.exists(DB_FILE):
        return []
    with open(DB_FILE, "r", encoding="utf-8") as f:
        return json.load(f)

def save_issues(issues):
    with open(DB_FILE, "w", encoding="utf-8") as f:
        json.dump(issues, f, ensure_ascii=False, indent=2)

def add_issue(description):
    issues = load_issues()
    new_issue = {
        "id": len(issues) + 1,
        "description": description,
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "is_completed": False
    }
    issues.append(new_issue)
    save_issues(issues)
    print(f"成功上傳問題 [ID: {new_issue['id']}]")

def list_issues():
    issues = load_issues()
    if not issues:
        print("目前沒有任何問題紀錄。")
        return
    print("\n--- 問題清單 ---")
    for issue in issues:
        status = "✅ 已完成" if issue["is_completed"] else "❌ 未處理"
        print(f"[{issue['id']}] {issue['timestamp']} | {status}")
        print(f"    描述: {issue['description']}")
    print("----------------\n")

if __name__ == "__main__":
    import sys
    if len(sys.argv) < 2:
        print("用法: python issue_tracker.py [add|list] [描述]")
    elif sys.argv[1] == "add" and len(sys.argv) > 2:
        add_issue(sys.argv[2])
    elif sys.argv[1] == "list":
        list_issues()
