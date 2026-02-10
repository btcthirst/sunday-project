import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import DashboardHome from '../views/DashboardHome.vue'
import CitizenList from '../views/CitizenList.vue'
import CitizenForm from '../views/CitizenForm.vue'
import RegistrationList from '../views/RegistrationList.vue'
import Placeholder from '../views/Placeholder.vue'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: '/login'
    },
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: { requiresAuth: false }
    },
    {
        path: '/dashboard',
        component: Dashboard,
        meta: { requiresAuth: true },
        children: [
            {
                path: '',
                name: 'DashboardHome',
                component: DashboardHome
            },
            {
                path: 'citizens',
                name: 'Citizens',
                component: CitizenList
            },
            {
                path: 'citizens/new',
                name: 'NewCitizen',
                component: CitizenForm
            },
            {
                path: 'citizens/:id',
                name: 'EditCitizen',
                component: CitizenForm
            },
            {
                path: 'registrations',
                name: 'Registrations',
                component: RegistrationList
            },
            {
                path: 'audit',
                name: 'Audit',
                component: () => import('../views/AuditLogs.vue')
            },
            {
                path: 'help',
                name: 'Help',
                component: () => import('../views/Help.vue')
            },
            {
                path: 'settings',
                name: 'Settings',
                component: () => import('../views/Settings.vue')
            }
        ]
    }
]

const router = createRouter({
    history: createWebHashHistory(),
    routes
})

import { IsAuthenticated } from '../../wailsjs/go/main/App'

router.beforeEach(async (to, from, next) => {
    if (to.meta.requiresAuth) {
        try {
            const auth = await IsAuthenticated()
            if (!auth) {
                next({ name: 'Login' })
                return
            }
        } catch (e) {
            next({ name: 'Login' })
            return
        }
    }
    next()
})

export default router
