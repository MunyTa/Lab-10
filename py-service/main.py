from fastapi import FastAPI

app = FastAPI()

@app.get("/api/users/{user_id}")
async def get_user(user_id: str):
    return {"error": "not implemented"}

@app.post("/api/users")
async def create_user():
    return {"error": "not implemented"}

@app.get("/api/health")
async def health():
    return {"error": "not implemented"}