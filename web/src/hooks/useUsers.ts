import { useState, useEffect } from 'react'
import { usersAPI } from '../api/users'

export const useUsers = () => {
  const [users, setUsers] = useState<any[]>([])
  const [stats, setStats] = useState<any>(null)
  const [roles, setRoles] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filters, setFilters] = useState({
    search: '',
    role: 'all',
    status: 'all'
  })

  useEffect(() => {
    fetchUsers()
    fetchUserStats()
    fetchUserRoles()
  }, [filters])

  const fetchUsers = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const usersData = await usersAPI.getUsers(filters)
      setUsers(usersData)
      
      // Calculate stats from users data
      const calculatedStats = {
        total_users: usersData?.length || 0,
        active_users: usersData?.filter((user: any) => user.status === 'active')?.length || 0,
        inactive_users: usersData?.filter((user: any) => user.status === 'inactive')?.length || 0,
        by_role: usersData?.reduce((acc: any, user: any) => {
          acc[user.role] = (acc[user.role] || 0) + 1
          return acc
        }, {}) || {},
        by_status: usersData?.reduce((acc: any, user: any) => {
          acc[user.status] = (acc[user.status] || 0) + 1
          return acc
        }, {}) || {}
      }
      setStats(calculatedStats)
      
      // Extract unique roles from users data
      const uniqueRoles = [...new Set(usersData?.map((user: any) => user.role) || [])]
      setRoles(uniqueRoles)
      
    } catch (error) {
      setError('Failed to fetch users')
      console.error('Users fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchUserStats = async () => {
    // Stats are now calculated in fetchUsers to avoid broken endpoint
  }

  const fetchUserRoles = async () => {
    // Roles are now extracted in fetchUsers to avoid broken endpoint
  }

  const createUser = async (userData: any) => {
    try {
      await usersAPI.createUser(userData)
      await fetchUsers()
      await fetchUserStats()
    } catch (error) {
      setError('Failed to create user')
      throw error
    }
  }

  const updateUser = async (id: any, userData: any) => {
    try {
      await usersAPI.updateUser(id, userData)
      await fetchUsers()
    } catch (error) {
      setError('Failed to update user')
      throw error
    }
  }

  const deleteUser = async (id: any) => {
    try {
      await usersAPI.deleteUser(id)
      await fetchUsers()
      await fetchUserStats()
    } catch (error) {
      setError('Failed to delete user')
      throw error
    }
  }

  const getUserDetails = async (id: any) => {
    try {
      return await usersAPI.getUser(id)
    } catch (error) {
      setError('Failed to fetch user details')
      throw error
    }
  }

  const updateFilters = (newFilters: any) => {
    setFilters(prev => ({ ...prev, ...newFilters }))
  }

  const refreshData = () => {
    fetchUsers()
    fetchUserStats()
    fetchUserRoles()
  }

  return {
    users,
    stats,
    roles,
    loading,
    error,
    filters,
    createUser,
    updateUser,
    deleteUser,
    getUserDetails,
    updateFilters,
    refreshData
  }
}

export default useUsers
