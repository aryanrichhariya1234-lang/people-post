"use server";
const BASE_URL = process.env.NEXT_PUBLIC_API_URL;
export const getId = async () => {
  try {
    const res = await fetch(`${BASE_URL}/users/me`, {
      method: "GET",
      credentials: "include",
    });

    const data = await res.json();

    return data?.user?._id || null;
  } catch (err) {
    return null;
  }
};
