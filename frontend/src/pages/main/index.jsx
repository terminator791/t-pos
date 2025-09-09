import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import History from "@/components/partials/widget/chart/history";
import RadarChart from "@/components/partials/widget/chart/radar-chart";
import Earnings from "@/components/partials/widget/chart/earnings";
import RecentOrderTable from "@/components/partials/Table/order-table";

const country = [
  {
    name: "Usa",
    flag: Usa,
    count: "$6.41",
    icon: "heroicons:arrow-small-up",
  },
  {
    name: "Brazil",
    flag: Brasil,
    count: "$2.33",
    icon: "heroicons:arrow-small-up",
  },
  {
    name: "Japan",
    flag: Japan,
    count: "$7.12",
    icon: "heroicons:arrow-small-down",
  },
  {
    name: "Italy",
    flag: Italy,
    count: "$754",
    icon: "heroicons:arrow-small-down",
  },
  {
    name: "India",
    flag: India,
    count: "$699",
    icon: "heroicons:arrow-small-up",
  },
  {
    name: "India",
    flag: India,
    count: "$624",
    icon: "heroicons:arrow-small-up",
  },
];
const source = [
  {
    name: "Direct Source",
    flag: "ph:circle-half",
    count: "1.2k",
    icon: "heroicons:arrow-small-down",
  },
  {
    name: "Social Network",
    flag: "ph:share-network",
    count: "0.33k",
    icon: "heroicons:arrow-small-down",
  },
  {
    name: "Email Newsletter",
    flag: "ph:chat-text",
    count: "31.12k",
    icon: "heroicons:arrow-small-up",
  },
  {
    name: "Referrals",
    flag: "ph:arrow-square-out",
    count: "890",
    icon: "heroicons:arrow-small-down",
  },
  {
    name: "ADVT",
    flag: "ph:percent",
    count: "765",
    icon: "heroicons:arrow-small-up",
  },
  {
    name: "Other",
    flag: "ph:star-four",
    count: "3.4k",
    icon: "heroicons:arrow-small-up",
  },
];

const Dashboard = () => {
  return (
    <div className=" space-y-5">
      <div className="grid xl:grid-cols-4 sm:grid-cols-2 grid-cols-1 gap-5">
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Products</div>
              <div className="flex-none">
                <div className="h-10 w-10  rounded-full bg-indigo-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:package" />
                </div>
              </div>
            </div>
            <div>
              <span className=" text-2xl font-medium text-gray-800  dark:text-white">
                145
              </span>
              <span className="  space-x-2 block mt-4 ">
                <span className="badge bg-indigo-500/10 text-indigo-500 ">
                  8.2%
                </span>
                <span className=" text-sm text-gray-500 dark:text-gray-400">
                  Since last week
                </span>
              </span>
            </div>
          </div>
        </Card>
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Licenses</div>
              <div className="flex-none">
                <div className="h-10 w-10  rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:certificate" />
                </div>
              </div>
            </div>
            <div>
              <span className=" text-2xl font-medium text-gray-800  dark:text-white">
                24
              </span>
              <span className="  space-x-2 block mt-4 ">
                <span className="badge bg-yellow-500/10 text-yellow-500 ">
                  2.1%
                </span>
                <span className=" text-sm text-gray-500 dark:text-gray-400">
                  Active licenses
                </span>
              </span>
            </div>
          </div>
        </Card>
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Customers</div>
              <div className="flex-none">
                <div className="h-10 w-10  rounded-full bg-red-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:users" />
                </div>
              </div>
            </div>
            <div>
              <span className=" text-2xl font-medium text-gray-800  dark:text-white">
                1,421
              </span>
              <span className="  space-x-2 block mt-4 ">
                <span className="badge bg-red-500/10 text-red-500 ">12.5%</span>
                <span className=" text-sm text-gray-500 dark:text-gray-400">
                  Total customers
                </span>
              </span>
            </div>
          </div>
        </Card>
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Users</div>
              <div className="flex-none">
                <div className="h-10 w-10  rounded-full bg-green-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:user-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className=" text-2xl font-medium text-gray-800  dark:text-white">
                18
              </span>
              <span className="  space-x-2 block mt-4 ">
                <span className="badge bg-green-500/10 text-green-500 ">
                  Active
                </span>
                <span className=" text-sm text-gray-500 dark:text-gray-400">
                  System users
                </span>
              </span>
            </div>
          </div>
        </Card>
      </div>
      {/* end grid */}
      <div className="grid   grid-cols-12 gap-5">
        <div className="xl:col-span-7 col-span-12">
          <Card
            className="!bg-indigo-500 "
            title="Sales History"
            titleClass="!text-white"
            noborder
          >
            <History />
          </Card>
        </div>
        <div className="xl:col-span-5 col-span-12">
          <Card title="System Overview" subscribe>
            <RadarChart />

            <div className="grid  grid-cols-3 gap-2 py-4">
              <div className=" text-center">
                <div>
                  <div className="h-10 w-10 mb-2 mx-auto rounded-md bg-indigo-500/10 text-indigo-500 text-xl flex items-center justify-center">
                    <Icon icon="ph:package" />
                  </div>
                </div>
                <div>
                  <div className=" font-medium text-gray-800 dark:text-white text-sm truncate mb-[2px]">
                    Active Products
                  </div>
                  <div className=" text-xs text-gray-400">142</div>
                </div>
              </div>
              {/* end single */}
              <div className=" text-center">
                <div>
                  <div className="h-10 w-10 mb-2 mx-auto rounded-md bg-green-500/10 text-green-500 text-xl flex items-center justify-center">
                    <Icon icon="ph:users" />
                  </div>
                </div>
                <div>
                  <div className=" font-medium text-gray-800 dark:text-white text-sm truncate mb-[2px]">
                    Active Customers
                  </div>
                  <div className=" text-xs text-gray-400">1,398</div>
                </div>
              </div>
              {/* end single */}
              <div className=" text-center">
                <div>
                  <div className="h-10 w-10 mb-2 mx-auto rounded-md bg-yellow-500/10 text-yellow-500 text-xl flex items-center justify-center">
                    <Icon icon="ph:certificate" />
                  </div>
                </div>
                <div>
                  <div className=" font-medium text-gray-800 dark:text-white text-sm truncate mb-[2px]">
                    Valid Licenses
                  </div>
                  <div className=" text-xs text-gray-400">22</div>
                </div>
              </div>
              {/* end single */}
            </div>
            {/* end support ticket */}
          </Card>
        </div>
      </div>
      {/* end grid */}
      <div className="grid xl:grid-cols-2 gap-5 ">
        <Card title="Recent Activity" subtitle="Last 30 days overview">
          <div className="space-y-4">
            <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="h-8 w-8 bg-green-500 rounded-full flex items-center justify-center">
                  <Icon icon="ph:plus" className="text-white text-sm" />
                </div>
                <div>
                  <div className="text-sm font-medium">New Product Added</div>
                  <div className="text-xs text-gray-500">Product X was added to inventory</div>
                </div>
              </div>
              <div className="text-xs text-gray-400">2 hours ago</div>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="h-8 w-8 bg-blue-500 rounded-full flex items-center justify-center">
                  <Icon icon="ph:user-plus" className="text-white text-sm" />
                </div>
                <div>
                  <div className="text-sm font-medium">New Customer Registration</div>
                  <div className="text-xs text-gray-500">John Smith joined the platform</div>
                </div>
              </div>
              <div className="text-xs text-gray-400">5 hours ago</div>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="h-8 w-8 bg-yellow-500 rounded-full flex items-center justify-center">
                  <Icon icon="ph:certificate" className="text-white text-sm" />
                </div>
                <div>
                  <div className="text-sm font-medium">License Renewed</div>
                  <div className="text-xs text-gray-500">Enterprise license extended</div>
                </div>
              </div>
              <div className="text-xs text-gray-400">1 day ago</div>
            </div>
          </div>
        </Card>
        
        <Card title="Quick Stats">
          <Earnings />
        </Card>
      </div>
      {/* end grid */}
      <div>
        <div className="card-title mb-5">Recent Orders</div>
        <RecentOrderTable />
      </div>
    </div>
  );
};

export default Dashboard;
