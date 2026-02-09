import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import RegistrationList from './RegistrationList.vue'
import { ListRegistrations } from '../../wailsjs/go/main/App'
import { NDataTable, NInput, NRadioGroup, NRadioButton, NTag, NIcon, NCard, NSpace } from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    ListRegistrations: vi.fn(),
}))

// Mock vue-router
const pushMock = vi.fn()
vi.mock('vue-router', () => ({
    useRouter: () => ({
        push: pushMock
    })
}))

// Mock naive-ui composables
vi.mock('naive-ui', async (importOriginal) => {
    const actual = await importOriginal() as any
    return {
        ...actual,
        useMessage: () => ({
            error: vi.fn(),
            success: vi.fn()
        })
    }
})

describe('RegistrationList.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        vi.useFakeTimers()
    })

    const mountOptions = {
        global: {
            components: {
                NDataTable,
                NInput,
                NRadioGroup,
                NRadioButton,
                NTag,
                NIcon,
                NCard,
                NSpace
            }
        }
    }

    it('loads registrations on mount', async () => {
        ; (ListRegistrations as any).mockResolvedValue({
            items: [
                { id: 1, citizen_name: 'John Doe', registration_date: '2023-01-01', is_active: true, settlement: 'Kyiv', street: 'Baseina', house_number: '1' }
            ],
            total: 1,
            total_pages: 1
        })

        const wrapper = mount(RegistrationList, mountOptions)

        await wrapper.vm.$nextTick()

        expect(ListRegistrations).toHaveBeenCalledWith('', true, 1, 20)
        expect((wrapper.vm as any).registrations).toHaveLength(1)
        expect((wrapper.vm as any).registrations[0].citizen_name).toBe('John Doe')
    })

    it('searches with debouncing', async () => {
        ; (ListRegistrations as any).mockResolvedValue({ items: [], total: 0, total_pages: 0 })
        const wrapper = mount(RegistrationList, mountOptions)

        const input = wrapper.findComponent(NInput)
        await input.vm.$emit('update:value', 'search query')

        // Should not be called immediately because of debounce
        expect(ListRegistrations).toHaveBeenCalledTimes(1) // Initial call on mount

        vi.advanceTimersByTime(500)

        expect(ListRegistrations).toHaveBeenCalledTimes(2)
        expect(ListRegistrations).toHaveBeenLastCalledWith('search query', true, 1, 20)
    })

    it('filters registrations by status', async () => {
        ; (ListRegistrations as any).mockResolvedValue({ items: [], total: 0, total_pages: 0 })
        const wrapper = mount(RegistrationList, mountOptions)

        const radioGroup = wrapper.findComponent(NRadioGroup)

        // Filter "Inactive" (Зняті)
        await radioGroup.vm.$emit('update:value', 'inactive')
        expect(ListRegistrations).toHaveBeenLastCalledWith('', false, 1, 20)

        // Filter "All"
        await radioGroup.vm.$emit('update:value', 'all')
        expect(ListRegistrations).toHaveBeenLastCalledWith('', null, 1, 20)
    })
})
