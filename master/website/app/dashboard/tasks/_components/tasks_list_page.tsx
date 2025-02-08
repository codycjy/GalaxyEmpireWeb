'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Switch } from "@/components/ui/switch";
import EditTaskDialog from './edit_task_dialog';
import { toast } from 'sonner';
import { useSearchParams } from 'next/navigation';
import CreateTaskDialog from './create_task_dialog';


interface Target {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  galaxy: number;
  system: number;
  planet: number;
  is_moon: boolean;
  task_id: number;
}


interface Task {
  ID: number;            // 大写 ID
  CreatedAt: string;     // 大写 CreatedAt
  UpdatedAt: string;     // 大写 UpdatedAt
  DeletedAt: string | null;
  name: string;
  next_start: string;
  enabled: boolean;
  account_id: number;
  task_type: number;
  targets: Target[];
  repeat: number;
  next_index: number;
  target_num: number;
  start_planet_id: number;  // 添加这个字段
  start_planet: Target;
  fleet: {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    lf: number;
    hf: number;
    cr: number;
    bs: number;
    dr: number;
    de: number;
    ds: number;
    bomb: number;
    guard: number;
    satellite: number;
    cargo: number;
    task_id: number;
  };
}

interface AccountResponse {
  succeed: boolean;
  data: {
    id: number;
    username: string;
    email: string;
    server: string;
    ExpireAt: string;
    tasks: Task[];
  };
  traceID: string;
}

export default function TaskListPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const searchParams = useSearchParams();
  const accountId = searchParams.get('accountId');
  const [accountInfo, setAccountInfo] = useState<{
    username: string;
    password: string;
    server: string;
    email: string;
  } | null>(null);
  
  // 获取任务类型名称
  const getTaskTypeName = (type: number) => {
    switch (type) {
      case 1:
        return '攻击';
      case 4:
        return '探索';
      default:
        return '未知';
    }
  };

  // 格式化时间戳
  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  // 格式化目标显示
  const formatTarget = (target: Target | null) => {
    if (!target) return '未设置';
    return `[${target.galaxy}:${target.system}:${target.planet}]`;
  };

  const fetchTasks = useCallback(async () => {
    if (!accountId) {
      toast.error('账号ID不能为空');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }

      const taskResponse = await fetch(`/api/v1/account/${accountId}`, {
        headers: {
          'Authorization': token,
        },
      });

      if (!taskResponse.ok) {
        throw new Error('获取任务列表失败');
      }

      const responseData = await taskResponse.json() as AccountResponse;
      console.log('API Response:', responseData.data.tasks);

      if (responseData.succeed) {
        console.log(' data:', responseData.data);
        const tasks = responseData.data.tasks;
        console.log('Tasks data:', tasks); // 检查任务数据
        setTasks(tasks);
        
        setAccountInfo({
          username: responseData.data.username,
          password: '', 
          server: responseData.data.server,
          email: responseData.data.email
        });
      } else {
        toast.error('获取任务列表失败');
      }

    } catch (error) {
      console.error('获取数据失败:', error);
      toast.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [accountId]);

  useEffect(() => {
    if (accountId) {
      fetchTasks();
    }
  }, [accountId, fetchTasks]);

  // 添加任务成功后的回调
  const handleTaskCreated = useCallback(() => {
    fetchTasks();
  }, [fetchTasks]);

  if (!accountId) {
    return (
      <div className="p-6">
        <div className="text-center text-gray-500">
          请先选择一个账号
        </div>
      </div>
    );
  }
  
  const handleDeleteTask = async (taskId: number) => {
    if (!confirm('确定要删除此任务吗？')) {
      return;
    }
  
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }
  
      const response = await fetch('/api/v1/task', {
        method: 'DELETE',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          id: taskId
        })
      });
  
      if (!response.ok) {
        throw new Error('删除失败');
      }
  
      toast.success('删除成功');
      fetchTasks(); // 刷新任务列表
    } catch (error) {
      console.error('删除失败:', error);
      toast.error('删除失败');
    }
  };

  return (
    <div className="p-6">
    <div className="flex justify-between items-center mb-6">
      <h1 className="text-2xl font-bold">任务列表</h1>
      <div className="flex gap-4">
        {accountId && (
          <CreateTaskDialog 
            accountId={accountId} 
            onSuccess={handleTaskCreated}
          />
        )}
        <Button onClick={() => fetchTasks()} disabled={loading}>
          {loading ? '刷新中...' : '刷新'}
        </Button>
      </div>
    </div>

    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="text-center">任务名称</TableHead>
            <TableHead className="text-center">类型</TableHead>
            <TableHead className="text-center">目标星球</TableHead>
            <TableHead className="text-center">重复次数</TableHead>
            <TableHead className="text-center">编辑任务</TableHead>
            <TableHead className="text-center">启用状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-4">
                加载中...
              </TableCell>
            </TableRow>
          ) : tasks.length > 0 ? (
            tasks.map((task) => (
              <TableRow key={task.ID}>
                <TableCell className="text-center">{task.name}</TableCell>
                <TableCell className="text-center">{getTaskTypeName(task.task_type)}</TableCell>
                <TableCell className="text-center">
                  {task.targets && task.targets.length > 0 ? (
                    task.targets.map((target, index) => (
                      <span key={target.ID}>
                        {index > 0 ? ', ' : ''}
                        [{target.galaxy}:{target.system}:{target.planet}]
                      </span>
                    ))
                  ) : (
                    '未设置'
                  )}
                </TableCell>
                <TableCell className="text-center">{task.repeat}</TableCell>
                <TableCell className="text-center">
                  <div className="flex justify-center gap-2">
                    <EditTaskDialog 
                      task={task} 
                      onSuccess={fetchTasks}
                    />
                    <Button 
                      variant="outline" 
                      size="sm"
                      className="text-red-500 hover:text-red-700"
                      onClick={() => handleDeleteTask(task.ID)}
                    >
                      删除
                    </Button>
                  </div>
                </TableCell>
                <TableCell className="text-center">
                  <Switch
                    checked={task.enabled}
                    onCheckedChange={async (checked: boolean) => {
                      try {
                        const token = localStorage.getItem('token');
                        const response = await fetch(`/api/v1/task/${task.ID}/enabled`, {
                          method: 'PUT',
                          headers: token ? {
                            'Authorization': token,
                            'Content-Type': 'application/json',
                          } : undefined,
                          body: JSON.stringify({ enabled: checked })
                        });
  
                        if (!response.ok) {
                          throw new Error('更新失败');
                        }
  
                        toast.success('更新成功');
                        fetchTasks(); // 刷新任务列表
                      } catch (error) {
                        console.error('更新失败:', error);
                        toast.error('更新失败');
                      }
                    }}
                  />
                </TableCell>
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-4">
                暂无任务数据
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
    </div>
  );
}