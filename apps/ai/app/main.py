import sys, os
from concurrent import futures

# chat_pb2_grpc.py (generated) does a flat `import chat_pb2`, so gen/ must be
# on sys.path before any other import touches it.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "gen"))

import grpc
import chat_pb2
import chat_pb2_grpc



class ChatService(chat_pb2_grpc.ChatServiceServicer):
    def Send(self, request, context):
        print(f"[py-ml] received: {request.message}")
        return chat_pb2.ChatResponse(reply=f"py-ml got it: {request.message}")


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    chat_pb2_grpc.add_ChatServiceServicer_to_server(ChatService(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    print("[py-ml] gRPC server listening on :50051")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()