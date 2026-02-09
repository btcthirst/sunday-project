import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CitizenForm from './CitizenForm.vue'
import {
    CreateCitizen,
    GetCitizen,
    UpdateCitizen,
    SearchCitizens,
    GetFamilyMembers
} from '../../wailsjs/go/main/App'
import {
    NTabs, NTabPane, NForm, NFormItem, NInput, NSelect,
    NDatePicker, NRadioGroup, NRadio, NButton, NCard, NGrid, NGi, NSpace
} from 'naive-ui'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    CreateCitizen: vi.fn(),
    GetCitizen: vi.fn(),
    UpdateCitizen: vi.fn(),
    SearchCitizens: vi.fn(),
    GetFamilyMembers: vi.fn(),
    GenerateCitizenCertificate: vi.fn(),
}))

// Mock vue-router
const pushMock = vi.fn()
const backMock = vi.fn()
const routeMock = { params: {} }

vi.mock('vue-router', () => ({
    useRoute: () => routeMock,
    useRouter: () => ({
        push: pushMock,
        back: backMock
    })
}))

// Mock naive-ui composables
vi.mock('naive-ui', async (importOriginal) => {
    const actual = await importOriginal() as any
    return {
        ...actual,
        useMessage: () => ({
            error: vi.fn(),
            success: vi.fn(),
            warning: vi.fn()
        })
    }
})

describe('CitizenForm.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        routeMock.params = {}
    })

    const mountOptions = {
        global: {
            components: {
                NTabs, NTabPane, NForm, NFormItem, NInput, NSelect,
                NDatePicker, NRadioGroup, NRadio, NButton, NCard, NGrid, NGi, NSpace
            },
            stubs: {
                // Stub complex sub-components that aren't the focus of this test
                'registration-history': true,
                'certificate-options-modal': true
            }
        }
    }

    it('renders create form by default', async () => {
        const wrapper = mount(CitizenForm, mountOptions)
        expect(wrapper.text()).toContain('Новий громадянин')
        expect(GetCitizen).not.toHaveBeenCalled()
    })

    it('loads data in edit mode', async () => {
        routeMock.params = { id: '1' }
        const mockCitizen = {
            id: 1,
            last_name: 'Іванов',
            first_name: 'Іван',
            birth_date: '1990-01-01',
            passport_type: 'old',
            gender: 'M'
        }
            ; (GetCitizen as any).mockResolvedValue(mockCitizen)
            ; (GetFamilyMembers as any).mockResolvedValue([])

        const wrapper = mount(CitizenForm, mountOptions)

        // Wait for onMounted
        await new Promise(resolve => setTimeout(resolve, 0))
        await wrapper.vm.$nextTick()

        expect(wrapper.text()).toContain('Редагування громадянина')
        expect(GetCitizen).toHaveBeenCalledWith(1)
        expect((wrapper.vm as any).formValue.last_name).toBe('Іванов')
    })

    it('validates required fields on save', async () => {
        const wrapper = mount(CitizenForm, mountOptions)
        const vm = wrapper.vm as any

        // Clear defaults to trigger validation
        vm.formValue.last_name = ''

        await vm.handleSave()

        expect(CreateCitizen).not.toHaveBeenCalled()
    })

    it('switches tabs correctly', async () => {
        const wrapper = mount(CitizenForm, mountOptions)
        const vm = wrapper.vm as any

        expect(vm.activeTab).toBe('personal')

        vm.activeTab = 'documents'
        await wrapper.vm.$nextTick()

        expect(wrapper.text()).toContain('Зразок паспорта')
    })

    it('adds a family member after search', async () => {
        const wrapper = mount(CitizenForm, mountOptions)
        const vm = wrapper.vm as any

        vm.activeTab = 'family'
        await wrapper.vm.$nextTick()

            // Simulate search
            ; (SearchCitizens as any).mockResolvedValue([
                { id: 2, full_name: 'Петров Петро', birth_date: '1985-05-05' }
            ])

        await vm.handleSearchMembers('Петров')
        expect(vm.memberOptions).toHaveLength(1)

        // Simulate selection and adding
        vm.selectedMemberID = 2
        vm.selectedRelation = 'Брат'
        // Need to populate memberOptions correctly for addRelation to find it
        vm.memberOptions = [{ value: 2, full_info: { full_name: 'Петров Петро' } }]

        vm.addRelation()

        expect(vm.familyRelations).toHaveLength(1)
        expect(vm.familyRelations[0].full_name).toBe('Петров Петро')
    })
})
