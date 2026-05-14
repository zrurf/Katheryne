import { ChatSidebar } from "../components/chat/ChatSidebar";
import { ChatView } from "../components/chat/ChatView";

export function ChatPage() {
  return (
    <div class="h-screen flex overflow-hidden">
      <ChatSidebar />
      <ChatView />
    </div>
  );
}