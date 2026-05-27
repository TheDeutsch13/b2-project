import { axiosInstance } from "./axiosInstance";

export interface SupportMessage {
  id: number;
  thread_id: number;
  sender_id: number;
  sender_role: "user" | "staff";
  sender_name: string;
  body: string;
  created_at: string;
}

export interface SupportThread {
  id: number;
  user_id: number;
  user_email: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface SupportThreadView {
  thread: SupportThread;
  messages: SupportMessage[];
}

export interface SupportThreadListItem extends SupportThread {
  last_message_body: string;
  last_message_at?: string;
  message_count: number;
}

export const supportApi = {
  getMyThread: async (): Promise<SupportThreadView> => {
    const response = await axiosInstance.get<SupportThreadView>("/api/support/my");
    return response.data;
  },

  sendMyMessage: async (
    body: string,
    senderName: string
  ): Promise<SupportMessage> => {
    const response = await axiosInstance.post<SupportMessage>(
      "/api/support/my/messages",
      { body, sender_name: senderName }
    );
    return response.data;
  },

  listThreads: async (
    includeClosed = false
  ): Promise<SupportThreadListItem[]> => {
    const response = await axiosInstance.get<SupportThreadListItem[]>(
      "/api/support/threads",
      { params: includeClosed ? { include_closed: "1" } : undefined }
    );
    return response.data;
  },

  closeThread: async (threadId: number): Promise<SupportThread> => {
    const response = await axiosInstance.patch<SupportThread>(
      `/api/support/threads/${threadId}/close`
    );
    return response.data;
  },

  deleteThread: async (threadId: number): Promise<void> => {
    await axiosInstance.delete(`/api/support/threads/${threadId}`);
  },

  getThread: async (threadId: number): Promise<SupportThreadView> => {
    const response = await axiosInstance.get<SupportThreadView>(
      `/api/support/threads/${threadId}`
    );
    return response.data;
  },

  sendStaffMessage: async (
    threadId: number,
    body: string,
    senderName: string
  ): Promise<SupportMessage> => {
    const response = await axiosInstance.post<SupportMessage>(
      `/api/support/threads/${threadId}/messages`,
      { body, sender_name: senderName }
    );
    return response.data;
  },
};
