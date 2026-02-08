import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CitizenList from './CitizenList.vue'
import { NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    ListCitizens: vi.fn(),
    SearchCitizens: vi.fn(),
    DeleteCitizen: vi.fn(),
    RestoreCitizen: vi.fn(),
}))

// Mock Wails models
vi.mock('../../wailsjs/go/models', () => ({
    models: {
        CitizenOutput: class {
            id = 0
            full_name = ''
            birth_date = ''
            passport_series = ''
            passport_number = ''
            passport_masked = ''
            tax_number = ''
            gender = ''
            phone = ''
            email = ''
            active_address = ''
            deleted = false
            constructor(source: any = {}) {
                Object.assign(this, source);
            }
        }
    }
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
        }),
        useDialog: () => ({
            warning: vi.fn()
        })
    }
})

import { ListCitizens } from '../../wailsjs/go/main/App'

describe('CitizenList.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('loads and displays citizens on mount', async () => {
        // Setup mock return value
        const mockCitizens = [
            { id: 1, full_name: 'John Doe', passport_masked: 'AA 123***', deleted: false }
        ]
            ; (ListCitizens as any).mockResolvedValue({
                items: mockCitizens,
                total: 1,
                page: 1,
                total_pages: 1
            })

        const wrapper = mount(CitizenList, {
            global: {
                stubs: {
                    // Stub naive-ui components to avoid complex rendering in tests
                    // But we want to check if data is passed to NDataTable
                    // 'n-data-table': true 
                    // If we stub it, we can't check contents easily unless we inspect props.
                },
                components: {
                    NConfigProvider,
                    NMessageProvider,
                    NDialogProvider
                }
            }
        })

        // Wait for promises to resolve
        await new Promise(resolve => setTimeout(resolve, 100))
        await wrapper.vm.$nextTick()

        expect(ListCitizens).toHaveBeenCalled()

        // Check if data is loaded into component state
        // We can check internal state or rendered output if we didn't fully stub
        // Since naive-ui components are complex, checking internal state is easier sometimes
        // But testing-library style is better.
        // Let's check if the text "John Doe" appears.
        // Note: virtual table might not render rows if height is 0 in happy-dom.
        // So checking ListCitizens call is a good first step.
        // Also check wrapper.vm.citizens
        expect((wrapper.vm as any).citizens).toHaveLength(1)
        expect((wrapper.vm as any).citizens[0].full_name).toBe('John Doe')
    })

    it('navigates to create page on button click', async () => {
        const wrapper = mount(CitizenList)

        const addButton = wrapper.findAll('.n-button').find(b => b.text().includes('Додати громадянина'))
        expect(addButton?.exists()).toBe(true)

        await addButton?.trigger('click')
        expect(pushMock).toHaveBeenCalledWith({ name: 'NewCitizen' })
    })
})
