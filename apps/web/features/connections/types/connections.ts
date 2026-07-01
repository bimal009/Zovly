import { ConnectedPage } from "@repo/types";


export type FacebookConnectionStatus = {
  connected: boolean;
  pages: ConnectedPage[];
};

export type InstagramConnectionStatus = {
  connected: boolean;
  facebook_linked: boolean;
  account: ConnectedPage | null;
};


