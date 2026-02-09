import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AuditLogs from './AuditLogs.vue'
import { GetAuditLogs } from '../../wailsjs/go/main/App'
import { NDataTable, NTag, NButton, NCard, NSpace } from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    GetAuditLogs: vi.fn(),
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

describe('AuditLogs.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    const mountOptions = {
        global: {
            components: {
                NDataTable,
                NTag,
                NButton,
                NCard,
                NSpace
            }
        }
    }

    it('loads audit logs on mount', async () => {
        const mockLogs = [
            { timestamp: '2023-01-01T12:00:00Z', operator_name: 'admin', action_type: 'CREATE', table_name: 'citizens', description: 'Created citizen' }
        ]
            ; (GetAuditLogs as any).mockResolvedValue(mockLogs)

        const wrapper = mount(AuditLogs, mountOptions)

        await wrapper.vm.$nextTick()

        expect(GetAuditLogs).toHaveBeenCalledWith(500)
        expect((wrapper.vm as any).logs).toHaveLength(1)
        expect((wrapper.vm as any).logs[0].operator_name).toBe('admin')
    })

    it('refreshes logs on button click', async () => {
        ; (GetAuditLogs as any).mockResolvedValue([])
        const wrapper = mount(AuditLogs, mountOptions)

        await wrapper.vm.$nextTick()
        expect(GetAuditLogs).toHaveBeenCalledTimes(1)

        const refreshButton = wrapper.findComponent(NButton)
        await refreshButton.trigger('click')

        expect(GetAuditLogs).toHaveBeenCalledTimes(2)
    })
})
