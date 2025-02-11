'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Plus, Minus, X, ChevronDown } from 'lucide-react';
import { toast } from 'sonner';

interface Target {
  galaxy: number;
  system: number;
  planet: number;
  is_moon: boolean;
}

interface Fleet {
  lf: number;         // 轻型战斗机
  hf: number;         // 重型战斗机
  cr: number;         // 巡洋舰
  bs: number;         // 战列舰
  dr: number;         // 无畏舰
  de: number;         // 驱逐舰
  ds: number;         // 死星
  bomb: number;       // 轰炸机
  guard: number;      // 守卫者
  satellite: number;  // 卫星
  cargo: number;      // 大型运输船
}

interface CreateTaskDialogProps {
    accountId: string;
    onSuccess: () => void;
  }

interface PlanetCheckResponse {
    uuid: string;
    trace_id: string;
}

interface PlanetQueryResult {
    // TODO: 根据实际返回数据结构定义   
    planet_id: number;
    succeed: boolean;
    trace_id: string;
}

export default function CreateTaskDialog({ accountId, onSuccess }: CreateTaskDialogProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [taskName, setTaskName] = useState('');
  const [taskType, setTaskType] = useState('1');
  const [repeat, setRepeat] = useState(1);
  const [targets, setTargets] = useState<Target[]>([]);
  const [startPlanet, setStartPlanet] = useState<Target>({
    galaxy: 0,
    system: 0,
    planet: 0,
    is_moon: false
  });
  const [newTarget, setNewTarget] = useState<Target>({
    galaxy: 0,
    system: 0,
    planet: 0,
    is_moon: false
  });
  const [fleet, setFleet] = useState<Fleet>({
    lf: 0,
    hf: 0,
    cr: 0,
    bs: 0,
    dr: 0,
    de: 0,
    ds: 0,
    bomb: 0,
    guard: 0,
    satellite: 0,
    cargo: 0
  });

  const fleetConfig = [
    { key: 'lf', name: '轻型战斗机' },
    { key: 'hf', name: '重型战斗机' },
    { key: 'cr', name: '巡洋舰' },
    { key: 'bs', name: '战列舰' },
    { key: 'dr', name: '无畏舰' },
    { key: 'de', name: '驱逐舰' },
    { key: 'ds', name: '死星' },
    { key: 'bomb', name: '轰炸机' },
    { key: 'guard', name: '守卫者' },
    { key: 'satellite', name: '卫星' },
    { key: 'cargo', name: '大型运输船' }
  ];

  const [isChecking, setIsChecking] = useState(false);
  const [startPlanetId, setStartPlanetId] = useState<number>(0);
  const [checkUuid, setCheckUuid] = useState<string>('');
  

  const handleFleetChange = (key: keyof Fleet, increment: boolean) => {
    setFleet(prev => ({
      ...prev,
      [key]: Math.max(0, prev[key] + (increment ? 1 : -1))
    }));
  };

  const handleFleetInputChange = (key: keyof Fleet, value: string) => {
    const numValue = parseInt(value) || 0;
    if (numValue >= 0) {
      setFleet(prev => ({
        ...prev,
        [key]: numValue
      }));
    }
  };

  const handleAddTarget = () => {
    if (newTarget.galaxy && newTarget.system && newTarget.planet) {
      setTargets([...targets, { ...newTarget }]);
      setNewTarget({ galaxy: 0, system: 0, planet: 0, is_moon: false });
    } else {
      toast.error('请填写完整的目标坐标');
    }
  };

  const handleDeleteTarget = (index: number) => {
    setTargets(targets.filter((_, i) => i !== index));
  };

  const initializeFleet = () => {
    setFleet({
      lf: 0,
      hf: 0,
      cr: 0,
      bs: 0,
      dr: 0,
      de: 0,
      ds: 0,
      bomb: 0,
      guard: 0,
      satellite: 0,
      cargo: 0
    });
  };

  const manualCheckResult = async () => {
    if (!checkUuid) {
      toast.error('没有可查询的检查ID');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }

      const resultResponse = await fetch(`/api/v1/task/planet/${checkUuid}`, {
        headers: {
          'Authorization': token,
        },
      });
      
      if (resultResponse.ok) {
        const planetData = await resultResponse.json() as PlanetQueryResult;
        if (planetData.succeed && planetData.planet_id > 0) {
          setStartPlanetId(planetData.planet_id);
          toast.success('星球检查完成');
        } else {
          toast.error('星球检查失败，请重试');
        }
      } else {
        toast.error('查询失败，请稍后重试');
      }
    } catch (error) {
      console.error('查询失败:', error);
      toast.error('查询失败');
    }
  };

  const checkStartPlanet = async () => {
    if (!startPlanet.galaxy || !startPlanet.system || !startPlanet.planet) {
      toast.error('请填写完整的起始星球坐标');
      return;
    }

    if (!accountId) {
      toast.error('请先选择账号');
      return;
    }

    setIsChecking(true);
    const checkingToast = toast.loading('正在检查星球信息...');

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }

      // 第一步：发起检查请求
      const checkResponse = await fetch('/api/v1/task/planet/query', {
        method: 'POST',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          target: {
            galaxy: startPlanet.galaxy,
            system: startPlanet.system,
            planet: startPlanet.planet,
            is_moon: startPlanet.is_moon
          },
          account: {
            id: parseInt(accountId)
          }
        }),
      });

      if (!checkResponse.ok) {
        throw new Error('发起检查请求失败');
      }

      const checkData = await checkResponse.json() as PlanetCheckResponse;
      setCheckUuid(checkData.uuid);
      
      // 第二步：轮询获取结果
      let attempts = 0;
      const maxAttempts = 5; // 最多轮询30次
      const pollInterval = 500; // 1秒间隔
      // TODO: 
      const pollResult = async (uuid: string): Promise<void> => {
        try {
          if (attempts >= maxAttempts) {
            throw new Error('检查超时');
          }

          attempts++;
          const resultResponse = await fetch(`/api/v1/task/planet/${uuid}`, {
            method: 'GET',
            headers: {
              'Authorization': token,
            },
          });
          console.log(resultResponse);
          
          if (resultResponse.ok) {
              const planetData = await resultResponse.json() as PlanetQueryResult;
              if (planetData.succeed && planetData.planet_id > 0) {
                  setStartPlanetId(planetData.planet_id);
                  toast.success('星球检查完成');
                  return;
              } else if (planetData.succeed === false) {
                  throw new Error('星球检查失败');
              }
          } else {
              // 不管是 404 还是 500，都继续轮询
              await new Promise(resolve => setTimeout(resolve, pollInterval));
              return pollResult(uuid);
          }
        } catch (error) {
          if (attempts < maxAttempts) {
              // 如果还没达到最大尝试次数，继续轮询
              await new Promise(resolve => setTimeout(resolve, pollInterval));
              return pollResult(uuid);
          }
          throw error;
      }
      };

      // 开始轮询
      await pollResult(checkData.uuid);
      
    } catch (error) {
      console.error('检查星球失败:', error);
      setStartPlanetId(0);
      toast.error(error instanceof Error ? error.message : '检查星球失败');
    } finally {
      toast.dismiss(checkingToast);
      setIsChecking(false);
    }
  };

  const handleSubmit = async () => {
    if (!taskName) {
      toast.error('请输入任务名称');
      return;
    }

    if (!startPlanet.galaxy || !startPlanet.system || !startPlanet.planet) {
      toast.error('请设置起始星球坐标');
      return;
    }

    // 添加对 startPlanetId 的检查
    if (startPlanetId === 0) {
      toast.error('请先检查起始星球并确保检查成功');
      return;
    }

    if (targets.length === 0) {
      toast.error('请至少添加一个目标');
      return;
    }

    const hasShips = Object.values(fleet).some(count => count > 0);
    if (!hasShips) {
      toast.error('请至少配置一种舰船');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录或登录已过期');
        return;
      }

      // 构造请求体
      const taskData = {
        name: taskName,
        task_type: parseInt(taskType),
        start_planet_id: startPlanetId,
        start_planet: { 
          galaxy: startPlanet.galaxy,
          system: startPlanet.system,
          planet: startPlanet.planet,
          is_moon: startPlanet.is_moon
        },
        targets: targets.map(target => ({
          galaxy: target.galaxy,
          system: target.system,
          planet: target.planet,
          is_moon: target.is_moon
        })),
        fleet: {
          bomb: fleet.bomb,
          bs: fleet.bs,
          cargo: fleet.cargo,
          cr: fleet.cr,
          de: fleet.de,
          dr: fleet.dr,
          ds: fleet.ds,
          guard: fleet.guard,
          hf: fleet.hf,
          lf: fleet.lf,
          satellite: fleet.satellite,
        },
        repeat: repeat,
        enabled: true,
        status: "ready",
        next_index: 0,
        next_start: 0,
        target_num: targets.length,
        account_id: parseInt(accountId)
      };

      console.log('Creating task with data:', taskData); // 添加日志

      const response = await fetch('/api/v1/task', {
        method: 'POST',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(taskData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || '创建任务失败');
      }

      const data = await response.json();
      
      if (data.succeed) {
        toast.success('任务创建成功');
        setIsOpen(false);
        onSuccess();
        // 重置表单
        setTaskName('');
        setTaskType('1');
        setStartPlanet({ galaxy: 0, system: 0, planet: 0, is_moon: false });
        setStartPlanetId(0);  // 重置星球ID
        setTargets([]);
        setFleet({
          lf: 0, hf: 0, cr: 0, bs: 0, dr: 0, de: 0,
          ds: 0, bomb: 0, guard: 0, satellite: 0, cargo: 0
        });
      } else {
        toast.error(data.message || '创建任务失败');
      }
    } catch (error) {
      console.error('创建任务失败:', error);
      toast.error(error instanceof Error ? error.message : '创建任务失败');
    }
  };

  // 监听起始星球坐标变化
  const handleStartPlanetChange = (field: keyof Target, value: number) => {
    setStartPlanet(prev => ({
      ...prev,
      [field]: value
    }));
    // 重置检查状态
    setStartPlanetId(0);
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          新建任务
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader className="sticky top-0 bg-white z-10 pb-4 border-b">
            <DialogTitle>新建任务</DialogTitle>
        </DialogHeader>
        
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label>任务名称</Label>
            <Input
              value={taskName}
              onChange={(e) => setTaskName(e.target.value)}
              placeholder="请输入任务名称"
            />
          </div>

          <div className="grid gap-2">
            <Label>任务类型</Label>
            <Select value={taskType} onValueChange={setTaskType}>
              <SelectTrigger>
                <SelectValue placeholder="选择任务类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">攻击</SelectItem>
                <SelectItem value="4">探索</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-2">
            <Label>起始星球</Label>
            <div className="flex gap-2">
              <Input
                type="number"
                placeholder="galaxy"
                value={startPlanet.galaxy || ''}
                onChange={(e) => handleStartPlanetChange('galaxy', parseInt(e.target.value) || 0)}
                className="text-center"
              />
              <Input
                type="number"
                placeholder="system"
                value={startPlanet.system || ''}
                onChange={(e) => handleStartPlanetChange('system', parseInt(e.target.value) || 0)}
                className="text-center"
              />
              <Input
                type="number"
                placeholder="planet"
                value={startPlanet.planet || ''}
                onChange={(e) => handleStartPlanetChange('planet', parseInt(e.target.value) || 0)}
                className="text-center"
              />
            </div>
            <div className="flex gap-2 items-center">
              <Button 
                onClick={checkStartPlanet}
                disabled={isChecking}
                className="w-20"
              >
                {isChecking ? '检查中...' : '检查'}
              </Button>
              {checkUuid && (
                <Button
                  variant="outline"
                  onClick={manualCheckResult}
                  disabled={isChecking || startPlanetId !== 0}
                  className="w-20"
                >
                  更新
                </Button>
              )}
              <div className="ml-2">
                {startPlanetId !== 0 ? (
                  <span className="text-green-600">✓ 检查成功</span>
                ) : checkUuid ? (
                  <span className="text-yellow-600">⟳ 等待检查结果</span>
                ) : null}
              </div>
            </div>
          </div>

          <div className="grid gap-2">
            <Label>重复次数</Label>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="icon"
                onClick={() => setRepeat(prev => Math.max(1, prev - 1))}
              >
                <Minus className="h-4 w-4" />
              </Button>
              <Input
                type="number"
                value={repeat}
                onChange={(e) => {
                  const value = parseInt(e.target.value);
                  if (value >= 1) setRepeat(value);
                }}
                min="1"
                className="text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                
              />
              <Button
                variant="outline"
                size="icon"
                onClick={() => setRepeat(prev => prev + 1)}
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <div className="grid gap-2">
            <Label>添加目标</Label>
            <div className="flex gap-2">
              <Input
                type="number"
                placeholder="galaxy"
                value={newTarget.galaxy || ''}
                onChange={(e) => setNewTarget(prev => ({
                  ...prev,
                  galaxy: parseInt(e.target.value) || 0
                }))}
                className="text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
              />
              <Input
                type="number"
                placeholder="system"
                value={newTarget.system || ''}
                onChange={(e) => setNewTarget(prev => ({
                  ...prev,
                  system: parseInt(e.target.value) || 0
                }))}
                className="text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
              />
              <Input
                type="number"
                placeholder="planet"
                value={newTarget.planet || ''}
                onChange={(e) => setNewTarget(prev => ({
                  ...prev,
                  planet: parseInt(e.target.value) || 0
                }))}
                className="text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
              />
              <Button onClick={handleAddTarget}>+</Button>
            </div>
          </div>

          <div className="border rounded-md">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="text-center">galaxy</TableHead>
                  <TableHead className="text-center">system</TableHead>
                  <TableHead className="text-center">planet</TableHead>
                  <TableHead className="w-[100px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {targets.length > 0 ? (
                  targets.map((target, index) => (
                    <TableRow key={index}>
                      <TableCell className="text-center">{target.galaxy}</TableCell>
                      <TableCell className="text-center">{target.system}</TableCell>
                      <TableCell className="text-center">{target.planet}</TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteTarget(index)}
                          className="text-red-500 hover:text-red-700"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center py-4 text-gray-500">
                      暂无目标数据
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        <Collapsible className="w-full">
          <CollapsibleTrigger className="flex w-full items-center justify-between p-4 border rounded-lg hover:bg-gray-50">
            <span className="font-medium">舰队配置</span>
            <ChevronDown className="h-4 w-4" />
          </CollapsibleTrigger>
          <CollapsibleContent className="p-4 space-y-4">
            <div className="flex justify-end">
              <Button 
                variant="outline" 
                onClick={initializeFleet}
                className="mb-2"
              >
                初始化舰队
              </Button>
            </div>
            <div className="grid grid-cols-2 gap-4">
              {fleetConfig.map(({ key, name }) => (
                <div key={key} className="flex items-center justify-between p-2 border rounded-lg">
                  <span className="text-sm">{name}</span>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => handleFleetChange(key as keyof Fleet, false)}
                    >
                      <Minus className="h-3 w-3" />
                    </Button>
                    <Input
                      type="number"
                      value={fleet[key as keyof Fleet]}
                      onChange={(e) => handleFleetInputChange(key as keyof Fleet, e.target.value)}
                      className="w-20 text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none" 
                      min="0"
                    />
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => handleFleetChange(key as keyof Fleet, true)}
                    >
                      <Plus className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CollapsibleContent>
        </Collapsible>

        <div className="flex justify-end gap-4">
          <Button variant="outline" onClick={() => setIsOpen(false)}>
            取消
          </Button>
          <Button onClick={handleSubmit}>
            创建
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}