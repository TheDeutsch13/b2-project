import { axiosInstance } from "./axiosInstance";

export interface CdekPickupPoint {
  code: string;
  name: string;
  address: string;
  city: string;
  work_time: string;
  phone: string;
  latitude: number;
  longitude: number;
}

export const cdekApi = {
  getPickupPoints: async (city: string): Promise<CdekPickupPoint[]> => {
    const response = await axiosInstance.get<CdekPickupPoint[]>("/api/cdek/points", {
      params: { city },
    });
    return response.data;
  },
};
