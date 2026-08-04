from fastapi import FastAPI, HTTPException
import os
import yaml
from pathlib import Path
from typing import List, Optional

app = FastAPI(title="Skill Management Server")

# 設定技能目錄路徑 (僅讀取伺服器目錄下的 skills)
SKILLS_DIR = Path(__file__).parent / "skills"

def extract_metadata(file_path: Path):
    """從 SKILL.md 提取 YAML Frontmatter"""
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
            if content.startswith("---"):
                parts = content.split("---")
                if len(parts) >= 3:
                    metadata = yaml.safe_load(parts[1])
                    metadata["id"] = file_path.parent.name
                    return metadata
    except Exception as e:
        print(f"解析 {file_path} 錯誤: {e}")
    return None

@app.get("/metadata")
def get_all_metadata():
    """取得所有 Skill 的元數據"""
    results = []
    if not SKILLS_DIR.exists():
         raise HTTPException(status_code=404, detail="Skills directory not found")
    
    for skill_path in SKILLS_DIR.glob("*/SKILL.md"):
        metadata = extract_metadata(skill_path)
        if metadata:
            results.append(metadata)
    return results

@app.get("/skill/{skill_id}")
def get_skill_content(skill_id: str):
    """根據 ID (資料夾名稱) 取得完整 Skill 內容，嚴格防止路徑穿越"""
    # 防止使用者傳入包含 '..' 或 '/' 的 skill_id
    if ".." in skill_id or "/" in skill_id or "\\" in skill_id:
        raise HTTPException(status_code=400, detail="非法路徑名稱")
    
    skill_file = (SKILLS_DIR / skill_id / "SKILL.md").resolve()
    
    # 確保解析後的絕對路徑確實位在 SKILLS_DIR 之下
    if not str(skill_file).startswith(str(SKILLS_DIR.resolve())):
        raise HTTPException(status_code=403, detail="存取權限不足：不得讀取上層目錄")

    if not skill_file.exists():
        raise HTTPException(status_code=404, detail=f"找不到技能 '{skill_id}'")
    
    with open(skill_file, "r", encoding="utf-8") as f:
        return {"id": skill_id, "content": f.read()}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=8000)
