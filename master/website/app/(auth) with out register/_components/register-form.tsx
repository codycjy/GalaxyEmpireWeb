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
        console.log('No captcha ID found');
        throw new Error('Captcha not loaded');
      }

      console.log('Sending login request with:', {
        username: data.username,
        captchaId: captcha.captchaId,
        userInput: data.captcha
      });

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

      console.log('Response status:', response.status);
      const result = await response.json();
      console.log('Full response data:', result);

      if (!response.ok) {
        const errorMessage = result.error || result.message || result.msg || 'Unknown error';
        console.log('Error message:', errorMessage);

        if (errorMessage.includes('record not found')) {
          toast.error('Login failed: Username does not exist or incorrect password, please try again!');
        } else if (errorMessage.includes('captcha')) {
          toast.error('Login failed: Invalid captcha');
        } else {
          toast.error(`Login failed: ${errorMessage}`);
        }
        getCaptcha();
        return;
      }

      if (result.token) {
        localStorage.setItem('token', result.token);
        localStorage.setItem('user', JSON.stringify(result.user));
        toast.success('Login successful!');
        router.push('/dashboard');
      } else {
        console.log('No token in successful response');
        toast.error('Login failed: Invalid response format');
        getCaptcha();
      }

    } catch (error) {
      console.error('Login error:', error);
      toast.error('Login failed: Please try again');
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
