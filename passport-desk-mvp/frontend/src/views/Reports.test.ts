import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Reports from './Reports.vue'
import { ExportCitizens } from '../../wailsjs/go/main/App'
import { NCard, NSpace, NAlert, NFormItem, NDatePicker, NButton, NIcon } from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    ExportCitizens: vi.fn(),
}))

// Mock naive-ui composables
vi.mock('naive-ui', async (importOriginal) => {
    const actual = await importOriginal() as any
    return {
        ...actual,
        useMessage: () => ({
            error: vi.fn(),
            success: vi.fn(),
            info: vi.fn()
        })
    }
})

describe('Reports.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    const mountOptions = {
        global: {
            components: {
                NCard,
                NSpace,
                NAlert,
                NFormItem,
                NDatePicker,
                NButton,
                NIcon
            }
        }
    }

    it('handles export without date range', async () => {
        ; (ExportCitizens as any).mockResolvedValue('/path/to/report.xlsx')
        const wrapper = mount(Reports, mountOptions)

        const exportButton = wrapper.findComponent(NButton)
        await exportButton.trigger('click')

        expect(ExportCitizens).toHaveBeenCalledWith('', '')
    })

    it('handles export with date range', async () => {
        ; (ExportCitizens as any).mockResolvedValue('/path/to/report.xlsx')
        const wrapper = mount(Reports, mountOptions)

            // Mock date range selection
            ; (wrapper.vm as any).dateRange = ['2023-01-01', '2023-12-31']

        const exportButton = wrapper.findComponent(NButton)
        await exportButton.trigger('click')

        expect(ExportCitizens).toHaveBeenCalledWith('2023-01-01', '2023-12-31')
    })

    it('shows loading state during export', async () => {
        let resolveExport: (val: string) => void = () => { }
        const exportPromise = new Promise<string>((resolve) => {
            resolveExport = resolve
        })
            ; (ExportCitizens as any).mockReturnValue(exportPromise)

        const wrapper = mount(Reports, mountOptions)
        const exportButton = wrapper.findComponent(NButton)

        await exportButton.trigger('click')
        expect((wrapper.vm as any).loading).toBe(true)

        resolveExport('/path/to/file.xlsx')
        await vi.waitFor(() => (wrapper.vm as any).loading === false)
        expect((wrapper.vm as any).loading).toBe(false)
    })
})
