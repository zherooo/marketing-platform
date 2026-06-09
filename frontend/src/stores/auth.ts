import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as apiLogin, register as apiRegister, getCurrentUser, changePassword as apiChangePassword } from '@/api'
import type { User, LoginRequest } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (newUser: User) => {
    user.value = newUser
  }

  const login = async (data: LoginRequest) => {
    const res = await apiLogin(data)
    setToken(res.token)
    setUser(res.user)
    return res
  }

  const register = async (data: LoginRequest & { email?: string; nickname?: string }) => {
    return apiRegister(data)
  }

  const fetchUser = async () => {
    if (!token.value) return null
    try {
      const res = await getCurrentUser()
      setUser(res.data)
      return res.data
    } catch {
      logout()
      return null
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string) => {
    return apiChangePassword(oldPassword, newPassword)
  }

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  const isLoggedIn = () => {
    return !!token.value
  }

  return {
    token,
    user,
    setToken,
    setUser,
    login,
    register,
    fetchUser,
    changePassword,
    logout,
    isLoggedIn
  }
})
