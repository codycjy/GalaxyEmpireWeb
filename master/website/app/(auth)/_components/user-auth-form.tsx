'use client';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import * as z from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { useState, useEffect } from 'react';

const API_BASE_URL = "/api/v1";  // 改回使用相对路径

const formSchema = z.object({
  username: z.string().min(1, { message: 'Username is required' }),
  password: z.string().min(6, { message: 'Password must be at least 6 characters' }),
  captcha: z.string().min(1, { message: 'Captcha input is required' }),
});

type UserFormValue = z.infer<typeof formSchema>;

interface CaptchaData {
  captchaId: string;
  captchaImg: string;
}

export default function UserAuthForm() {
  const [captcha, setCaptcha] = useState<CaptchaData | null>(null);
  const router = useRouter();
  
  const getCaptcha = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/captcha`, {
        method: 'GET',
        credentials: 'include',
      });
      
      if (!response.ok) throw new Error('Failed to fetch captcha');
      
      const data = await response.json();
      console.log('Captcha response:', data);
      
      if (data.succeed && data.captcha_id) {
        // 直接使用 captcha_id 构建图片 URL
        const captchaImgUrl = `${API_BASE_URL}/captcha/${data.captcha_id}`;
        
        setCaptcha({
          captchaId: data.captcha_id,
          captchaImg: captchaImgUrl
        });

        console.log('Captcha state:', {
          captchaId: data.captcha_id,
          ccaptchaImg: captchaImgUrl
        });
      }
      
     
    } catch (error) {
      console.error('Captcha error:', error);
      toast.error('Failed to load captcha');
    }
  };

  useEffect(() => {
    getCaptcha();
  }, []);

  const form = useForm<UserFormValue>({
    resolver: zodResolver(formSchema),
  });

  const onSubmit = async (data: UserFormValue) => {
    try {
      if (!captcha?.captchaId) {
        throw new Error('Captcha not loaded');
      }

      const response = await fetch(`${API_BASE_URL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'captchaId': captcha.captchaId,
          'userInput': data.captcha
        },
        body: JSON.stringify({
          username: data.username,
          password: data.password
        })
      });

      if (!response.ok) {
        throw new Error('Login failed');
      }

      const { token, user } = await response.json();

      // 存储 Token 和用户信息
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));

      toast.success('Login Successful!');
      router.push('/dashboard'); // 登录成功跳转
      // 弹出登录成功提示
      toast.success('Login Successful!');
    } catch (error) {
      toast.error('Login Failed');
      // 刷新验证码
      getCaptcha();
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        {/* Username */}
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input placeholder="Enter your username..." type="text" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Password */}
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input placeholder="Enter your password..." type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Captcha */}
        <FormField
          control={form.control}
          name="captcha"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Captcha</FormLabel>
              <div className="flex gap-2">
                <FormControl>
                  <Input placeholder="Enter captcha..." {...field} />
                </FormControl>
                <div className="relative h-10 w-24">
                  {captcha?.captchaImg && (
                    <img
                      src={`${captcha.captchaImg}`}
                      alt="captcha"
                      className="h-full w-full cursor-pointer object-contain"
                      onClick={() => getCaptcha()}
                    />
                  )}
                </div>
              </div>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" className="w-full">
          Login
        </Button>
      </form>
    </Form>
  );
}
