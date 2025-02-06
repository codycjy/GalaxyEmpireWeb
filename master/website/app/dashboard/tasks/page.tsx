import { Metadata } from 'next';
import TaskListPage from './_components/tasks_list_page';

export const metadata: Metadata = {
  title: 'Dashboard : Tasks'
};

// 确保返回一个 React 组件
export default function Page() {
  return (
    <div>
      <TaskListPage />
    </div>
  );
}