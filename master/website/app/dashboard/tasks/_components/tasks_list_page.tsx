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
import { toast } from 'sonner';
import { useSearchParams } from 'next/navigation';
import CreateTaskDialog from './create_task_dialog';

interface Target {
  id: number;
  task_id: number;
  galaxy: number;
  system: number;
  position: number;
}

interface Fleet {
  id: number;
  task_id: number;
  ships: Record<string, number>; // 船只配置
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
  targets: Target[] | null;
  repeat: number;
  next_index: number;
  target_num: number;
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
    return `[${target.galaxy}:${target.system}:${target.position}]`;
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

      const taskResponse = await fetch(`/api/v1/task/account/${accountId}`, {
        headers: {
          'Authorization': token,
        },
      });

      if (!taskResponse.ok) {
        throw new Error('获取任务列表失败');
      }

      const responseData = await taskResponse.json() as AccountResponse;
      
      if (responseData.succeed) {
        // 直接使用 data.tasks
        const tasks = responseData.data.tasks.map(task => ({
          id: task.ID,
          name: task.name,
          next_start: new Date(task.next_start).getTime() / 1000,
          enabled: task.enabled,
          account_id: task.account_id,
          task_type: task.task_type,
          status: task.status || 'pending',
          targets: task.targets || [],
          repeat: task.repeat,
          next_index: task.next_index,
          target_num: task.target_num,
          fleet: {
            id: task.fleet.ID,
            lf: task.fleet.lf,
            hf: task.fleet.hf,
            cr: task.fleet.cr,
            bs: task.fleet.bs,
            dr: task.fleet.dr,
            de: task.fleet.de,
            ds: task.fleet.ds,
            bomb: task.fleet.bomb,
            guard: task.fleet.guard,
            satellite: task.fleet.satellite,
            cargo: task.fleet.cargo,
            task_id: task.fleet.task_id,
            created_at: task.fleet.CreatedAt,
            updated_at: task.fleet.UpdatedAt
          },
          created_at: task.CreatedAt,
          updated_at: task.UpdatedAt
        }));
        
        setTasks(tasks);
        
        // 更新账号信息
        setAccountInfo({
          username: responseData.data.username,
          password: '', // 密码不在响应中
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
            刷新
          </Button>
        </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="text-center">任务名称</TableHead>
              <TableHead className="text-center">类型</TableHead>
              <TableHead className="text-center">状态</TableHead>
              <TableHead className="text-center">目标星球</TableHead>
              <TableHead className="text-center">重复次数</TableHead>
              <TableHead className="text-center">下次开始</TableHead>
              <TableHead className="text-center">启用状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-4">
                  加载中...
                </TableCell>
              </TableRow>
            ) : tasks.length > 0 ? (
              tasks.map((task) => (
                <TableRow key={task.id}>
                  <TableCell className="text-center">{task.name}</TableCell>
                  <TableCell className="text-center">{getTaskTypeName(task.task_type)}</TableCell>
                  <TableCell className="text-center">
                    <span className={`px-2 py-1 rounded-full text-sm ${
                      task.status === 'completed' 
                        ? 'bg-green-100 text-green-700'
                        : task.status === 'failed'
                        ? 'bg-red-100 text-red-700'
                        : 'bg-yellow-100 text-yellow-700'
                    }`}>
                      {task.status === 'completed' ? '已完成' 
                        : task.status === 'failed' ? '失败'
                        : task.status === 'ready' ? '就绪'
                        : '进行中'}
                    </span>
                  </TableCell>
                  <TableCell className="text-center">
                    {task.targets ? task.targets.map(formatTarget).join(', ') : '未设置'}
                  </TableCell>
                  <TableCell className="text-center">{task.repeat}</TableCell>
                  <TableCell className="text-center">
                    {task.next_start ? formatTimestamp(task.next_start) : '未设置'}
                  </TableCell>
                  <TableCell className="text-center">
                    <span className={`px-2 py-1 rounded-full text-sm ${
                      task.enabled 
                        ? 'bg-green-100 text-green-700'
                        : 'bg-gray-100 text-gray-700'
                    }`}>
                      {task.enabled ? '已启用' : '已禁用'}
                    </span>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-4">
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