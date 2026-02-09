import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Dashboard from './Dashboard.vue'
import { GetCurrentOperator, Logout, UpdateActivity } from '../../wailsjs/go/main/App'
import {
    NLayout, NLayoutSider, NLayoutHeader, NLayoutContent,
    NMenu, NBreadcrumb, NBreadcrumbItem, NDropdown, NButton, NIcon
} from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    GetCurrentOperator: vi.fn(),
    Logout: vi.fn(),
    UpdateActivity: vi.fn(),
}))

// Mock vue-router
const pushMock = vi.fn()
const routeMock = { path: '/dashboard' }
vi.mock('vue-router', () => ({
    useRouter: () => ({
        push: pushMock
    }),
    useRoute: () => routeMock,
    RouterView: { template: '<div />' }
}))

// Mock naive-ui components that might cause issues if not real
// (Usually not needed if imported and registered, but for complex icons sometimes)

describe('Dashboard.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        routeMock.path = '/dashboard'
    })

    const mountOptions = {
        global: {
            components: {
                NLayout,
                NLayoutSider,
                NLayoutHeader,
                NLayoutContent,
                NMenu,
                NBreadcrumb,
                NBreadcrumbItem,
                NDropdown,
                NButton,
                NIcon
            },
            stubs: {
                'router-view': true
            }
        }
    }

    it('loads and displays operator name on mount', async () => {
        ; (GetCurrentOperator as any).mockResolvedValue({ id: 1, full_name: 'Admin User' })
        const wrapper = mount(Dashboard, mountOptions)

        await wrapper.vm.$nextTick()
        await new Promise(resolve => setTimeout(resolve, 0))
        await wrapper.vm.$nextTick()

        expect(GetCurrentOperator).toHaveBeenCalled()
        expect(wrapper.text()).toContain('Admin User')
    })

    it('navigates when menu item is clicked', async () => {
        ; (GetCurrentOperator as any).mockResolvedValue({ id: 1, full_name: 'Admin User' })
        const wrapper = mount(Dashboard, mountOptions)

        const menu = wrapper.findComponent(NMenu)
        await menu.vm.$emit('update:value', 'citizens')

        expect(pushMock).toHaveBeenCalledWith({ name: 'Citizens' })
    })

    it('handles logout', async () => {
        ; (GetCurrentOperator as any).mockResolvedValue({ id: 1, full_name: 'Admin User' })
        const wrapper = mount(Dashboard, mountOptions)

        // Dropdown select is usually triggered via handleUserMenuSelect
        await (wrapper.vm as any).handleUserMenuSelect('logout')

        expect(Logout).toHaveBeenCalled()
        expect(pushMock).toHaveBeenCalledWith('/login')
    })

    it('syncs menu selection with current route', async () => {
        routeMock.path = '/citizens'
            ; (GetCurrentOperator as any).mockResolvedValue({ id: 1, full_name: 'Admin User' })
        const wrapper = mount(Dashboard, mountOptions)

        await wrapper.vm.$nextTick()
        expect((wrapper.vm as any).activeMenu).toBe('citizens')
    })
})
