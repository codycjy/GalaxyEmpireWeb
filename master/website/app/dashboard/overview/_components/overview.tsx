'use client';

import { Card } from '@/components/ui/card';

export default function OverviewPage() {
  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">概览</h2>
      </div>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card className="p-6">
          <h3 className="text-sm font-medium">总访问量</h3>
          <div className="mt-2 text-2xl font-bold">0</div>
        </Card>
        <Card className="p-6">
          <h3 className="text-sm font-medium">今日访问</h3>
          <div className="mt-2 text-2xl font-bold">0</div>
        </Card>
        <Card className="p-6">
          <h3 className="text-sm font-medium">活跃账号</h3>
          <div className="mt-2 text-2xl font-bold">0</div>
        </Card>
        <Card className="p-6">
          <h3 className="text-sm font-medium">收益</h3>
          <div className="mt-2 text-2xl font-bold">0%</div>
        </Card>
      </div>
    </div>
  );
}
