'use client';

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { useState, useEffect } from "react";
import { Plus, Minus, X, ChevronDown } from 'lucide-react';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

interface Target {
  ID: number;
  galaxy: number;
  system: number;
  planet: number;
  is_moon: boolean;
  task_id: number;
}

interface Task {
  ID: number;
  name: string;
  next_start: string;
  enabled: boolean;
  account_id: number;
  task_type: number;
  targets: Target[];
  repeat: number;
  next_index: number;
  target_num: number;
  fleet: {
    ID: number;
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
  start_planet_id: number;
  start_planet: Target;
}

interface EditTaskDialogProps {
  task: Task;
  onSuccess: () => void;
}

interface PlanetCheckResponse {
  uuid: string;
  trace_id: string;
}

interface PlanetQueryResult {
  planet_id: number;
  succeed: boolean;
  trace_id: string;
}

export default function EditTaskDialog({ task, onSuccess }: EditTaskDialogProps) {
  const [editTask, setEditTask] = useState<Task>(task);
  const [open, setOpen] = useState(false);
  const [isChecking, setIsChecking] = useState(false);
  const [startPlanetId, setStartPlanetId] = useState<number>(0);
  const [checkUuid, setCheckUuid] = useState<string>('');
  const [startPlanet, setStartPlanet] = useState({
    galaxy: 0,
    system: 0,
    planet: 0,
    is_moon: false
  });

  useEffect(() => {
    if (open) {
      setStartPlanetId(0);
      setCheckUuid('');
      setStartPlanet({
        galaxy: task.start_planet?.galaxy || 0,
        system: task.start_planet?.system || 0,
        planet: task.start_planet?.planet || 0,
        is_moon: task.start_planet?.is_moon || false
      });
      setEditTask(task);
    }
  }, [open, task]);

  const handleStartPlanetChange = (field: keyof typeof startPlanet, value: number | boolean) => {
    setStartPlanet(prev => ({
      ...prev,
      [field]: value
    }));
    setStartPlanetId(0);
    setCheckUuid('');
  };

  const handleTargetChange = (index: number, field: keyof Target, value: number) => {
    setEditTask(prev => {
      const newTargets = [...prev.targets];
      newTargets[index] = { ...newTargets[index], [field]: value };
      return { ...prev, targets: newTargets };
    });
  };

  const handleFleetChange = (key: string, value: number) => {
    setEditTask(prev => ({
      ...prev,
      fleet: { ...prev.fleet, [key]: value }
    }));
  };

  const checkStartPlanet = async () => {
    if (!startPlanet.galaxy || !startPlanet.system || !startPlanet.planet) {
      toast.error('请填写完整的起始星球坐标');
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

      const checkResponse = await fetch('/api/v1/task/planet/query', {
        method: 'POST',
        headers: {
          'Authorization': token,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          target: startPlanet,
          account: {
            id: editTask.account_id
          }
        }),
      });

      if (!checkResponse.ok) {
        throw new Error('发起检查请求失败');
      }

      const checkData = await checkResponse.json() as PlanetCheckResponse;
      setCheckUuid(checkData.uuid);
      
      let attempts = 0;
      const maxAttempts = 5;
      const pollInterval = 500;

      const pollResult = async (uuid: string): Promise<void> => {
        try {
          if (attempts >= maxAttempts) {
            throw new Error('检查超时');
          }

          attempts++;
          const resultResponse = await fetch(`/api/v1/task/planet/${uuid}`, {
            headers: {
              'Authorization': token,
            },
          });
          
          if (resultResponse.ok) {
            const planetData = await resultResponse.json() as PlanetQueryResult;
            if (planetData.succeed && planetData.planet_id > 0) {
              setStartPlanetId(planetData.planet_id);
              toast.success('星球检查完成');
              return;
            }
          }
          
          await new Promise(resolve => setTimeout(resolve, pollInterval));
          return pollResult(uuid);
        } catch (error) {
          if (attempts < maxAttempts) {
            await new Promise(resolve => setTimeout(resolve, pollInterval));
            return pollResult(uuid);
          }
          throw error;
        }
      };

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // 1. 基本字段检查
    if (!editTask.name.trim()) {
      toast.error('请输入任务名称');
      return;
    }

    if (!editTask.account_id) {
      toast.error('账号ID无效');
      return;
    }

    if (typeof editTask.task_type !== 'number' || editTask.task_type <= 0) {
      toast.error('任务类型无效');
      return;
    }

    if (typeof editTask.repeat !== 'number' || editTask.repeat <= 0) {
      toast.error('重复次数必须大于0');
      return;
    }

    // 2. 起始星球检查
    const hasNewStartPlanet = startPlanet.galaxy || startPlanet.system || startPlanet.planet;
    if (hasNewStartPlanet) {
      if (!startPlanet.galaxy || !startPlanet.system || !startPlanet.planet) {
        toast.error('请完整填写起始星球坐标');
        return;
      }
      if (startPlanetId === 0) {
        toast.error('请先验证起始星球');
        return;
      }
    }

    // 3. 目标星球检查
    if (!editTask.targets || editTask.targets.length < 1) {
      toast.error('请至少添加一个目标星球');
      return;
    }

    // 检查所有目标星球的坐标是否完整且有效
    const invalidTarget = editTask.targets.find(target => {
      const validGalaxy = typeof target.galaxy === 'number' && target.galaxy > 0;
      const validSystem = typeof target.system === 'number' && target.system > 0;
      const validPlanet = typeof target.planet === 'number' && target.planet > 0;
      return !validGalaxy || !validSystem || !validPlanet;
    });

    if (invalidTarget) {
      toast.error('存在无效的目标星球坐标');
      return;
    }

    // 4. 舰队配置检查
    const fleet = editTask.fleet;
    if (!fleet) {
      toast.error('舰队配置无效');
      return;
    }

    const hasShips = Object.entries(fleet).some(([key, value]) => {
      return key !== 'id' && 
             key !== 'task_id' && 
             typeof value === 'number' && 
             value > 0;
    });

    if (!hasShips) {
      toast.error('请至少配置一种舰船');
      return;
    }

    // 5. 准备提交数据
    try {
      const submitData = {
        name: editTask.name.trim(),
        enabled: editTask.enabled,
        task_type: editTask.task_type,
        start_planet_id: hasNewStartPlanet ? startPlanetId : task.start_planet_id,
        targets: editTask.targets.map(target => ({
          galaxy: target.galaxy,
          system: target.system,
          planet: target.planet,
          is_moon: target.is_moon || false
        })),
        repeat: editTask.repeat,
        fleet: {
          lf: editTask.fleet.lf || 0,
          hf: editTask.fleet.hf || 0,
          cr: editTask.fleet.cr || 0,
          bs: editTask.fleet.bs || 0,
          dr: editTask.fleet.dr || 0,
          de: editTask.fleet.de || 0,
          ds: editTask.fleet.ds || 0,
          bomb: editTask.fleet.bomb || 0,
          guard: editTask.fleet.guard || 0,
          satellite: editTask.fleet.satellite || 0,
          cargo: editTask.fleet.cargo || 0
        }
      };

      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/task/${editTask.ID}`, {
        method: 'PUT',
        headers: token ? {
          'Authorization': token,
          'Content-Type': 'application/json',
        } : undefined,
        body: JSON.stringify(submitData)
      });

      if (!response.ok) {
        throw new Error('更新失败');
      }

      toast.success('更新成功');
      onSuccess();
      setOpen(false);
    } catch (error) {
      console.error('更新失败:', error);
      toast.error('更新失败');
    }
  };

  const handleAddTarget = () => {
    setEditTask(prev => ({
      ...prev,
      targets: [
        ...prev.targets,
        {
          ID: 0, // 新目标的ID会由后端生成
          galaxy: 0,
          system: 0,
          planet: 0,
          is_moon: false,
          task_id: prev.ID
        }
      ]
    }));
  };

  const handleRemoveTarget = (index: number) => {
    setEditTask(prev => ({
      ...prev,
      targets: prev.targets.filter((_, i) => i !== index),
      target_num: prev.targets.length - 1
    }));
  };

  // 添加舰队配置映射
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

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">编辑</Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>编辑任务</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            {/* 任务名称 */}
            <div className="grid gap-2">
              <Label>任务名称</Label>
              <Input
                value={editTask.name}
                onChange={(e) => setEditTask(prev => ({...prev, name: e.target.value}))}
              />
            </div>
            
            {/* 起始星球 */}
            <div className="grid gap-2">
              <Label>起始星球</Label>
              <div className="flex gap-2">
                <Input
                  type="number"
                  placeholder="galaxy"
                  value={startPlanet.galaxy || task.start_planet?.galaxy || ''}
                  onChange={(e) => handleStartPlanetChange('galaxy', parseInt(e.target.value))}
                  className="w-24"
                />
                <Input
                  type="number"
                  placeholder="system"
                  value={startPlanet.system || task.start_planet?.system || ''}
                  onChange={(e) => handleStartPlanetChange('system', parseInt(e.target.value))}
                  className="w-24"
                />
                <Input
                  type="number"
                  placeholder="planet"
                  value={startPlanet.planet || task.start_planet?.planet || ''}
                  onChange={(e) => handleStartPlanetChange('planet', parseInt(e.target.value))}
                  className="w-24"
                />
                <div className="flex items-center gap-2">
                  <Button 
                    type="button"
                    onClick={checkStartPlanet}
                    disabled={isChecking}
                    className="w-[80px]"
                  >
                    {isChecking ? '检查中' : '检查'}
                  </Button>
                  {checkUuid && (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={manualCheckResult}
                      disabled={isChecking || startPlanetId !== 0}
                      className="w-[80px]"
                    >
                      刷新
                    </Button>
                  )}
                  {startPlanetId !== 0 && (
                    <span className="text-green-600 flex items-center">
                      <svg
                        viewBox="0 0 1024 1024"
                        fill="currentColor"
                        className="w-4 h-4 mr-1"
                      >
                        <path d="M912 190h-69.9c-9.8 0-19.1 4.5-25.1 12.2L404.7 724.5 207 474a32 32 0 0 0-25.1-12.2H112c-6.7 0-10.4 7.7-6.3 12.9l273.9 347c12.8 16.2 37.4 16.2 50.3 0l488.4-618.9c4.1-5.1.4-12.8-6.3-12.8z"/>
                      </svg>
                      检查成功
                    </span>
                  )}
                </div>
              </div>
            </div>
            
            {/* 重复次数 */}
            <div className="grid gap-2">
              <Label>重复次数</Label>
              <Input
                type="number"
                min="1"
                value={editTask.repeat}
                onChange={(e) => setEditTask(prev => ({
                  ...prev,
                  repeat: parseInt(e.target.value) || 1
                }))}
                className="w-24"
              />
            </div>
            
            {/* 目标星球 */}
            <div className="grid gap-2">
              <Label>目标星球</Label>
              {editTask.targets.map((target, index) => (
                <div key={target.ID} className="flex gap-2">
                  <Input
                    type="number"
                    placeholder="galaxy"
                    value={target.galaxy}
                    onChange={(e) => handleTargetChange(index, 'galaxy', parseInt(e.target.value))}
                    className="w-24"
                  />
                  <Input
                    type="number"
                    placeholder="system"
                    value={target.system}
                    onChange={(e) => handleTargetChange(index, 'system', parseInt(e.target.value))}
                    className="w-24"
                  />
                  <Input
                    type="number"
                    placeholder="planet"
                    value={target.planet}
                    onChange={(e) => handleTargetChange(index, 'planet', parseInt(e.target.value))}
                    className="w-24"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemoveTarget(index)}
                    className="text-red-500 hover:text-red-700"
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                onClick={handleAddTarget}
                className="w-full mt-2"
              >
                添加目标星球
              </Button>
            </div>

            {/* 舰队配置 */}
            <Collapsible className="w-full">
              <CollapsibleTrigger className="flex w-full items-center justify-between p-4 border rounded-lg hover:bg-gray-50">
                <span className="font-medium">舰队配置</span>
                <ChevronDown className="h-4 w-4" />
              </CollapsibleTrigger>
              <CollapsibleContent className="p-4">
                <div className="grid grid-cols-2 gap-4">
                  {fleetConfig.map(({ key, name }) => (
                    <div key={key} className="flex gap-2 items-center">
                      <Label className="w-24">{name}</Label>
                      <Input
                        type="number"
                        value={editTask.fleet[key as keyof typeof editTask.fleet]}
                        onChange={(e) => handleFleetChange(key, parseInt(e.target.value))}
                        className="w-24"
                      />
                    </div>
                  ))}
                </div>
              </CollapsibleContent>
            </Collapsible>
          </div>
          
          <DialogFooter>
            <Button type="submit">保存更改</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
} 