import { createBrowserRouter } from "react-router-dom";
import App from "./shell";
import BucketListPage from "../pages/BucketListPage";
import BucketDetailPage from "../pages/BucketDetailPage";
import CreateBucketPage from "../pages/CreateBucketPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <BucketListPage /> },
      { path: "buckets/:name", element: <BucketDetailPage /> },
      { path: "buckets/new", element: <CreateBucketPage /> },
    ],
  },
]);
