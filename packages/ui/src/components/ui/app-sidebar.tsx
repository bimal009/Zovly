"use client";

import * as React from "react";
import {
  IconBolt,
  IconBrandFacebook,
  IconBrandInstagram,
  IconBrandTiktok,
  IconBrandWhatsapp,
  IconCalendar,
  IconChartBar,
  IconDashboard,
  IconHelp,
  IconInbox,
  IconPackage,
  IconPhoto,
  IconSettings,
  IconSpeakerphone,
  IconTools,
  IconUsers,
} from "@tabler/icons-react";

import { NavDocuments } from "./nav-documents";
import { NavMain } from "./nav-main";
import { NavSecondary } from "./nav-secondary";
import { NavUser } from "./nav-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "./sidebar";

const data = {
  user: {
    name: "Alex Johnson",
    email: "alex@mybusiness.com",
    avatar: "/avatars/user.jpg",
  },
  navMain: [
    { title: "Overview", url: "#", icon: IconDashboard },
    { title: "Content Hub", url: "#", icon: IconPhoto },
    { title: "Inbox", url: "#", icon: IconInbox },
    { title: "Analytics", url: "#", icon: IconChartBar },
    { title: "Leads", url: "#", icon: IconUsers },
    { title: "Campaigns", url: "#", icon: IconSpeakerphone },
    { title: "Bookings", url: "#", icon: IconCalendar },
    { title: "Products", url: "#", icon: IconPackage },
    { title: "Services", url: "#", icon: IconTools },
  ],
  platforms: [
    { name: "Instagram", url: "#", icon: IconBrandInstagram },
    { name: "TikTok", url: "#", icon: IconBrandTiktok },
    { name: "Facebook", url: "#", icon: IconBrandFacebook },
    { name: "WhatsApp", url: "#", icon: IconBrandWhatsapp },
  ],
  navSecondary: [
    { title: "Settings", url: "#", icon: IconSettings },
    { title: "Help & Support", url: "#", icon: IconHelp },
  ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:p-1.5!"
            >
              <a href="#">
                <IconBolt className="size-5!" />
                <span className="text-base font-semibold">SocialOS</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavDocuments items={data.platforms} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  );
}
