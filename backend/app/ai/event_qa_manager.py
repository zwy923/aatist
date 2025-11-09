from dotenv import load_dotenv
load_dotenv()

from backend.app.ai.tools.event_scraper import fetch_aalto_events
from langchain_community.vectorstores import FAISS
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.runnables import RunnableParallel, RunnablePassthrough

def build_event_index():
    """爬取活动并建立向量数据库"""
    events = fetch_aalto_events()
    texts = [f"{e['title']} — {e['description']}" for e in events]
    metadatas = [{"title": e["title"], "url": e["link"]} for e in events]

    embeddings = OpenAIEmbeddings()
    vectorstore = FAISS.from_texts(texts, embeddings, metadatas=metadatas)
    vectorstore.save_local("backend/app/ai/vectorstore/faiss_index")
    print("✅ 向量索引已保存。")


def get_event_qa_chain():
    """使用 Runnable 组合问答链"""
    vectorstore = FAISS.load_local(
        "backend/app/ai/vectorstore/faiss_index",
        OpenAIEmbeddings(),
        allow_dangerous_deserialization=True  # ✅ 关键参数
    )
    retriever = vectorstore.as_retriever(search_kwargs={"k": 3})

    llm = ChatOpenAI(model="gpt-4o-mini", temperature=0.2)

    prompt = ChatPromptTemplate.from_template("""
    You are an AI assistant for Aalto University.
    Answer the user's question using the provided context.
    Context: {context}
    Question: {question}
    """)

    chain = (
        RunnableParallel({"context": retriever, "question": RunnablePassthrough()})
        | prompt
        | llm
    )

    return chain



if __name__ == "__main__":
    print("🔍 正在构建 Aalto Event 向量索引...")
    build_event_index()
    print("✅ 索引创建完成！可以调用 get_event_qa_chain() 进行问答。")
