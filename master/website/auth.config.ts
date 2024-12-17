import NextAuth from 'next-auth'
import { type NextAuthConfig } from 'next-auth'
import CredentialsProvider from 'next-auth/providers/credentials'

export const authConfig = {
  providers: [
    CredentialsProvider({
      credentials: {
        username: {
          type: 'text',
          label: 'Username'
        },
        password: {
          type: 'password',
          label: 'Password'
        },
        captcha: {
          type: 'text',
          label: 'Verification Code'
        }
      },
      async authorize(credentials) {
        try {
          // ... 你的验证逻辑
          return null
        } catch (error) {
          return null
        }
      }
    })
  ],
  callbacks: {
    async jwt({ token }) {
      return token
    },
    async session({ session }) {
      return session
    }
  },
  pages: {
    signIn: '/'
  }
} satisfies NextAuthConfig

export const { handlers, auth, signIn, signOut } = NextAuth(authConfig)
