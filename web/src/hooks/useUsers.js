import { useState, useEffect } from 'react'
import { usersAPI } from '../api/users'

export const useUsers = () => {
  const [users, setUsers] = useState([])
  const [stats, setStats] = useState(null)
  const [roles, setRoles] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
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
    } catch (error) {
      setError('Failed to fetch users')
      console.error('Users fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchUserStats = async () => {
    try {
      const statsData = await usersAPI.getUserStats()
      setStats(statsData)
    } catch (error) {
      console.error('User stats fetch error:', error)
    }
  }

  const fetchUserRoles = async () => {
    try {
      const rolesData = await usersAPI.getUserRoles()
      setRoles(rolesData)
    } catch (error) {
      console.error('User roles fetch error:', error)
    }
  }

  const createUser = async (userData) => {
    try {
      await usersAPI.createUser(userData)
      await fetchUsers()
      await fetchUserStats()
    } catch (error) {
      setError('Failed to create user')
      throw error
    }
  }

  const updateUser = async (id, userData) => {
    try {
      await usersAPI.updateUser(id, userData)
      await fetchUsers()
    } catch (error) {
      setError('Failed to update user')
      throw error
    }
  }

  const deleteUser = async (id) => {
    try {
      await usersAPI.deleteUser(id)
      await fetchUsers()
      await fetchUserStats()
    } catch (error) {
      setError('Failed to delete user')
      throw error
    }
  }

  const getUserDetails = async (id) => {
    try {
      return await usersAPI.getUser(id)
    } catch (error) {
      setError('Failed to fetch user details')
      throw error
    }
  }

  const updateFilters = (newFilters) => {
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
