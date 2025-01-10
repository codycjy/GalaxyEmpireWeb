'use client';

import { useState, useEffect, useCallback } from 'react';
import { useUser } from '@/components/providers/user-provider';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger, 
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Plus, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

interface GameAccount {
  id: number;
  username: string;
  server: string;
  ExpireAt: string;
  tasks: any[]; // 可以根据实际 tasks 的结构定义更具体的类型
  status: string;
}

interface CheckResponse {
  succeed: boolean;
  traceID: string;
  uuid: string;
}

interface ApiResponse {
  succeed: boolean;
  data: {
    id: number;
    username: string;
    accounts: GameAccount[];
    balance: number;
  };
  traceID: string;
}

interface CreateAccountForm {
  username: string;
  password: string;
  server: string;
}

export default function AccountListingPage() {
  const { user } = useUser();
  const [accounts, setAccounts] = useState<GameAccount[]>([]);
  const [mounted, setMounted] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isChecking, setIsChecking] = useState(false);
  const [checkUuid, setCheckUuid] = useState<string | null>(null);
  const [isVerified, setIsVerified] = useState(false);
  const [checkInterval, setCheckInterval] = useState<NodeJS.Timeout | null>(null);
  const [newAccount, setNewAccount] = useState<CreateAccountForm>({
    username: '',
    password: '',
    server: '',
  });


  // 组件挂载和卸载
  useEffect(() => {
    setMounted(true);
    return () => {
      setMounted(false);
      // 清除轮询
      if (checkInterval) {
        clearInterval(checkInterval);
      }
    };
  }, [checkInterval]);

  // 清理所有状态和请求的函数
  const cleanupCheckStatus = useCallback(() => {
    // 只在对话框关闭时才清理验证状态
    if (checkInterval) {
      clearInterval(checkInterval);
      setCheckInterval(null);
    }
  }, [checkInterval]);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      cleanupCheckStatus();
    };
  }, [cleanupCheckStatus]);

  // 完整重置函数（仅在关闭对话框时使用）
  const resetForm = useCallback(() => {
    // 先清理轮询
    if (checkInterval) {
      clearInterval(checkInterval);
      setCheckInterval(null);
    }
    
    // 重置所有状态
    setIsChecking(false);
    setIsVerified(false);
    setCheckUuid(null);
    setNewAccount({
      username: '',
      password: '',
      server: '',
    });
  }, [checkInterval]);

  // 开始轮询检查状态
  const startPolling = useCallback((uuid: string, token: string) => {
    // 清除之前的轮询
    if (checkInterval) {
      clearInterval(checkInterval);
    }
    let attempts = 0;
    const maxAttempts = 3; // 最大轮询次数

    const interval = setInterval(async () => {
      try {
        attempts++; // 增加尝试次数
        const response = await fetch(`/api/v1/account/check/${uuid}`, {
          method: 'GET',
          headers: {
            'Authorization': token,
            'Content-Type': 'application/json',
          },
        });

        const data = await response.json();
        
        if (data.succeed) {
          // 先清除定时器
          clearInterval(interval);
          setCheckInterval(null);
          
          // 使用批量更新来确保状态一致性
          Promise.resolve().then(() => {
            setIsVerified(true);
            setIsChecking(false);
            toast.success('账号验证成功');
          });
        } else if (attempts >= maxAttempts) {
          // 达到最大尝试次数
          clearInterval(interval);
          setCheckInterval(null);
          setIsChecking(false);
          toast.error('验证超时，请重试');
          
        } else {
          // 添加验证中的提示
          toast.info('正在验证账号...', {
            id: 'checking-status', // 使用固定 ID 避免重复提示
          });
        }
      } catch (error) {
        console.error('Check status error:', error);
        setIsChecking(false);
        toast.error('验证检查失败');
        clearInterval(interval);
        setCheckInterval(null);
      }
    }, 100);

    setCheckInterval(interval);
  }, [checkInterval]);


    // 添加状态监听以便调试
  useEffect(() => {
    console.log('isChecking 状态变化:', isChecking);
  }, [isChecking]);

  useEffect(() => {
    console.log('isVerified 状态变化:', isVerified);
  }, [isVerified]);


  const handleAccountCheck = async () => {
    try {
      setIsChecking(true);
      const token = localStorage.getItem('token');
      
      if (!token) {
        toast.error('未登录或登录已过期');
        setIsChecking(false);
        return;
      }

      if (!newAccount.username || !newAccount.password || !newAccount.server) {
        toast.error('请填写账号、密码和服务器');
        setIsChecking(false);
        return;
      }

      const response = await fetch('/api/v1/account/check', {
        method: 'POST',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: newAccount.username,
          server: newAccount.server,
          password: newAccount.password
        }),
      });

      const data: CheckResponse = await response.json();
      
      if (data.uuid) {
        setCheckUuid(data.uuid);
        // 确保在开始轮询时保持 isChecking 为 true
        startPolling(data.uuid, token);
        toast.info('正在验证账号...', {
          id: 'checking-status'
        });
      } else {
        setIsChecking(false); // 在失败时重置状态
        toast.error('验证请求失败');
      }
    } catch (error) {
      toast.error('验证失败');
    }finally {
      setIsCreating(false);
    }
  };


  // 添加创建账号的函数
  const handleCreateAccount = async () => {
    try {
      setIsCreating(true);
      const token = localStorage.getItem('token');
      
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }

      if (!newAccount.username || !newAccount.password || !newAccount.server) {
        toast.error('请填写账号、密码和服务器');
        return;
      }
      
      console.log('发送验证请求');
      const response = await fetch('/api/v1/account', {
        method: 'POST',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newAccount),
      });

      const data = await response.json();
      console.log('验证响应:', data); // 添加日志
      
      if (data.succeed) {
        toast.success('账号创建成功');
        setIsAddDialogOpen(false);
        setNewAccount({
          username: '',
          password: '',
          server: '',
        });
        fetchAccounts(); // 刷新账号列表
      } else {
        toast.error(data.message || '创建失败');
      }
    } catch (error) {
      console.error('Create account error:', error);
      toast.error('创建账号失败');
    } finally {
      setIsCreating(false);
    }
  };

  // 检查认证状态的函数
  const checkAuth = () => {
    const token = localStorage.getItem('token');
    if (!token) {
      toast.error('未登录或登录已过期');
      window.location.href = '/login';  // 使用 window.location.href 作为备选方案
      return false;
    }
    return token;
  };
  
  
    // 3. 获取账号列表
  const fetchAccounts = useCallback(async () => {
    try {
      const token = checkAuth();
      if (!token) return;

      const response = await fetch(`/api/v1/account/user/${user?.id}`, {
        method: 'GET',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      const data: ApiResponse = await response.json();
      
      if (data.succeed) {
        // 直接获取 accounts 数组
        setAccounts(data.data.accounts);
      } else {
        toast.error('获取账号列表失败');
      }
    } catch (error) {
      console.error('Fetch accounts error:', error);
      toast.error('获取账号列表失败');
    }
  },[user?.id]);

  // 检查过期的函数
  const checkExpired = (expireAt: string) => {
    const expireDate = new Date(expireAt);
    const now = new Date();
    return expireDate < now;
  };

    // 4. 只在组件挂载且有用户 ID 时获取数据
    useEffect(() => {
      if (mounted && user?.id) {
        fetchAccounts();
      }
    }, [mounted, user?.id, fetchAccounts]);
  
    // 如果组件未挂载，返回加载状态
    if (!mounted) {
      return (
        <div className="flex items-center justify-center min-h-screen">
          <div>加载中...</div>
        </div>
      );
    }
  // 添加删除账号的函数
  const handleDeleteAccount = async (accountId: number) => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }

      const response = await fetch('/api/v1/account', {
        method: 'DELETE',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          id: accountId
        })
      });

      const data = await response.json();
      
      if (data.succeed) {
        toast.success('账号删除成功');
        // 刷新账号列表
        fetchAccounts();
      } else {
        toast.error(data.message || '删除失败');
      }
    } catch (error) {
      console.error('删除账号错误:', error);
      toast.error('删除账号失败');
    }
  };

// 修改表格渲染部分

return (
  <div className="container mx-auto py-6">
    <div className="flex justify-between items-center mb-6">
      <h1 className="text-2xl font-bold">游戏账号管理</h1>
      <Dialog 
        open={isAddDialogOpen} 
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) {
            resetForm();
            toast.dismiss('checking-status'); // 关闭验证中的提示
          }
        }}
      >
        <DialogTrigger asChild>
          <Button className="bg-rose-200 hover:bg-rose-300">
            <Plus className="h-4 w-4 mr-2" />
            添加账号
          </Button>
        </DialogTrigger>
        <DialogContent onInteractOutside={(e) => {
          if (isChecking || isCreating) {
            e.preventDefault();
          }
        }}>
          <DialogHeader>
            <DialogTitle>添加新账号</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="username">账号</Label>
              <Input
                id="username"
                value={newAccount.username}
                onChange={(e) => setNewAccount(prev => ({
                  ...prev,
                  username: e.target.value
                }))}
                placeholder="请输入账号"
                // 在验证过程中或验证成功后禁用输入
                disabled={isChecking || isVerified}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">密码</Label>
              <Input
                id="password"
                type="password"
                value={newAccount.password}
                onChange={(e) => setNewAccount(prev => ({
                  ...prev,
                  password: e.target.value
                }))}
                placeholder="请输入密码"
                disabled={isChecking || isVerified}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="server">服务器</Label>
              <Input
                id="server"
                value={newAccount.server}
                onChange={(e) => setNewAccount(prev => ({
                  ...prev,
                  server: e.target.value
                }))}
                placeholder="请输入服务器"
                disabled={isChecking || isVerified}
              />
            </div>
          </div>
          <div className="flex justify-end gap-4">
            <Button
              variant="outline"
              onClick={() => {
                setIsAddDialogOpen(false);
                resetForm();
                toast.dismiss('checking-status');
              }}
              disabled={isChecking || isCreating}
            >
              取消
            </Button>
            <Button
              onClick={() => {
                console.log('验证开始'); // 添加这行来测试按钮点击
                handleAccountCheck();
              }}
              disabled={isChecking || isVerified}
              className={`${
                isChecking 
                  ? 'bg-yellow-500 hover:bg-yellow-600'
                  : isVerified 
                    ? 'bg-green-500 hover:bg-green-600'
                    : 'bg-blue-500 hover:bg-blue-600'
              } text-white`}
            >
              
              {isChecking ? (
                  <>
                    <div className="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-black mr-2" />
                    验证中...
                  </>
                ) : isVerified ? (
                  <>
                    <span className="mr-2">✓</span>
                    已验证
                  </>
                ) : '验证账号'}
            </Button>
            <Button
              onClick={() =>{
                handleCreateAccount();
                setIsAddDialogOpen(false);
                resetForm();
              }}
              disabled={!isVerified || isCreating}
              className={`${isVerified ? 'bg-blue-500 hover:bg-blue-600' : 'bg-gray-300'}`}
            >
              {isCreating ? (
                <>
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-black mr-2" />
                  创建中...
                </>
              ) : '创建'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>

    <div className="w-full rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="text-center">账号</TableHead>
            <TableHead className="text-center">服务器</TableHead>
            <TableHead className="text-center">状态</TableHead>
            <TableHead className="text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {accounts.length > 0 ? (
            accounts.map((account) => (
              <TableRow key={account.id}>
                <TableCell className="text-center">{account.username}</TableCell>
                <TableCell className="text-center">{account.server}</TableCell>
                <TableCell className="text-center">
                  <div className="flex justify-center items-center gap-2">
                    <span className={`px-2 py-1 rounded-full text-sm ${
                      checkExpired(account.ExpireAt) 
                        ? 'bg-red-100 text-red-700' 
                        : 'bg-green-100 text-green-700'
                    }`}>
                      {checkExpired(account.ExpireAt) ? '已过期' : '生效中'} 
                      ({new Date(account.ExpireAt).toLocaleDateString('zh-CN')})
                    </span>
                  </div>
                </TableCell>
                <TableCell className="text-center">
                  <Dialog>
                    <DialogTrigger asChild>
                      <Button variant="ghost" size="icon" className="text-red-500 hover:text-red-700">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>确认删除</DialogTitle>
                      </DialogHeader>
                      <div className="py-4">
                        <p>确定要删除账号 &quot;{account.username}&quot; 吗？此操作无法撤销。</p>
                      </div>
                      <div className="flex justify-end gap-4">
                        <DialogClose asChild>
                          <Button variant="outline">取消</Button>
                        </DialogClose>
                        <Button 
                          variant="destructive" 
                          onClick={() => {
                            handleDeleteAccount(account.id);
                          }}
                        >
                          删除
                        </Button>
                      </div>
                    </DialogContent>
                  </Dialog>
                </TableCell>
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={4} className="text-center py-4">
                暂无账号数据
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  </div>
);
}