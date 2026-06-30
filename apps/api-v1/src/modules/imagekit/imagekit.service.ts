import { ImageKit } from "@imagekit/nodejs";
import "dotenv/config";

const imagekit = new ImageKit({
  privateKey: process.env.IMAGEKIT_PRIVATE_KEY!,
});

export const getImageKitSignature = () => {
  return imagekit.helper.getAuthenticationParameters();
};

export default imagekit;
