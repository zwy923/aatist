from dotenv import load_dotenv
load_dotenv()
import sys, os
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../../..")))

from backend.app.ai.event_qa_manager import get_event_qa_chain

qa = get_event_qa_chain()
query = "how many events are there related to computer science?"
resp = qa.invoke(query)
print("\n🧩 Question:", query)
print("💬 Answer:\n", resp.content)