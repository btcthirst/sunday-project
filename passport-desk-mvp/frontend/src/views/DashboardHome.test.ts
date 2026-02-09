import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DashboardHome from './DashboardHome.vue'
import { GetCurrentOperator, GetDashboardStats } from '../../wailsjs/go/main/App'
import { NGrid, NGi, NCard, NStatistic, NIcon, NSpace, NButton, NH1 } from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    GetCurrentOperator: vi.fn(),
    GetDashboardStats: vi.fn(),
}))

// Mock vue-router
const pushMock = vi.fn()
vi.mock('vue-router', () => ({
    useRouter: () => ({
        push: pushMock
    })
}))

describe('DashboardHome.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    const mountOptions = {
        global: {
            components: {
                NGrid,
                NGi,
                NCard,
                NStatistic,
                NIcon,
                NSpace,
                NButton,
                NH1
            }
        }
    }

    it('loads and displays stats and operator name on mount', async () => {
        ; (GetCurrentOperator as any).mockResolvedValue({ full_name: 'Test Operator' })
            ; (GetDashboardStats as any).mockResolvedValue({
                total_citizens: 100,
                active_registrations: 50,
                new_this_month: 10
            })

        const wrapper = mount(DashboardHome, mountOptions)

        // Wait for onMounted
        await new Promise(resolve => setTimeout(resolve, 0))
        await wrapper.vm.$nextTick()

        expect(GetCurrentOperator).toHaveBeenCalled()
        expect(GetDashboardStats).toHaveBeenCalled()

        expect(wrapper.text()).toContain('Test Operator')
        expect(wrapper.text()).toContain('100')
        expect(wrapper.text()).toContain('50')
        expect(wrapper.text()).toContain('10')
    })

    it('navigates to NewCitizen on "Додати громадянина" click', async () => {
        ; (GetCurrentOperator as any).mockResolvedValue({ full_name: 'Test' })
            ; (GetDashboardStats as any).mockResolvedValue({})

        const wrapper = mount(DashboardHome, mountOptions)
        await new Promise(resolve => setTimeout(resolve, 0))

        const addButton = wrapper.findAllComponents(NButton).find(b => b.text().includes('Додати громадянина'))
        expect(addButton?.exists()).toBe(true)

        await addButton?.trigger('click')
        expect(pushMock).toHaveBeenCalledWith({ name: 'NewCitizen' })
    })

    it('navigates to Citizens on "Пошук" click', async () => {
        ; (GetCurrentOperator as any).mockResolvedValue({ full_name: 'Test' })
            ; (GetDashboardStats as any).mockResolvedValue({})

        const wrapper = mount(DashboardHome, mountOptions)
        await new Promise(resolve => setTimeout(resolve, 0))

        const searchButton = wrapper.findAllComponents(NButton).find(b => b.text().includes('Пошук'))
        expect(searchButton?.exists()).toBe(true)

        await searchButton?.trigger('click')
        expect(pushMock).toHaveBeenCalledWith({ name: 'Citizens' })
    })
})
