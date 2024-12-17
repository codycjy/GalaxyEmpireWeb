# Galaxy Empire Web

基于 Next.js 14 构建的 Galaxy Empire 网站项目。


## 项目结构
website/
├── app/ # Next.js 应用主目录
│ ├── api/ # API 路由
│ ├── auth/ # 认证相关页面
│ └── layout.tsx # 根布局
├── components/ # React 组件
├── constants/ # 常量定义
├── hooks/ # 自定义 React Hooks
├── lib/ # 工具函数和服务
├── public/ # 静态资源
└── styles/ # 全局样式

## 环境变量配置

在website目录下创建 `.env.local` 文件并配置以下环境变量：
NEXT_PUBLIC_API_URL=http://localhost:9333
NEXTAUTH_URL="http://localhost:3000"
