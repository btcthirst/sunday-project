import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Login from './Login.vue'
import {
    IsFirstRun,
    Login as LoginApi,
    SetupInitialOperator,
    RecoverByMasterKey
} from '../../wailsjs/go/main/App'

// Mock Wails backend functions
vi.mock('../../wailsjs/go/main/App', () => ({
    IsFirstRun: vi.fn(),
    Login: vi.fn(),
    SetupInitialOperator: vi.fn(),
    RecoverByMasterKey: vi.fn(),
}))

vi.mock('../../wailsjs/runtime', () => ({
    ClipboardSetText: vi.fn().mockResolvedValue(true),
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

import { NInput, NForm, NFormItem, NButton, NAlert, NModal, NIcon } from 'naive-ui'

describe('Login.vue', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    const mountOptions = {
        global: {
            components: {
                NInput,
                NForm,
                NFormItem,
                NButton,
                NAlert,
                NModal,
                NIcon
            }
        }
    }

    it('shows setup form on first run', async () => {
        ; (IsFirstRun as any).mockResolvedValue(true)
        const wrapper = mount(Login, mountOptions)

        // Wait for onMounted
        await new Promise(resolve => setTimeout(resolve, 0))
        await wrapper.vm.$nextTick()

        expect(wrapper.text()).toContain('Створіть обліковий запис оператора')
        expect(wrapper.findComponent(NInput).exists()).toBe(true)
        expect(wrapper.html()).toContain('Прізвище Ім\'я По-батькові')
    })

    it('shows login form on regular run', async () => {
        ; (IsFirstRun as any).mockResolvedValue(false)
        const wrapper = mount(Login, mountOptions)

        await new Promise(resolve => setTimeout(resolve, 0))
        await wrapper.vm.$nextTick()

        expect(wrapper.text()).toContain('Вхід в систему')
        expect(wrapper.html()).not.toContain('Прізвище Ім\'я По-батькові')
    })

    it('validates matching passwords during setup', async () => {
        ; (IsFirstRun as any).mockResolvedValue(true)
        const wrapper = mount(Login, mountOptions)
        await new Promise(resolve => setTimeout(resolve, 0))

        const vm = wrapper.vm as any
        vm.formData.username = 'admin'
        vm.formData.fullName = 'Admin User'
        vm.formData.password = 'pass123'
        vm.formData.confirmPassword = 'different'

        await vm.handleSubmit()

        expect(SetupInitialOperator).not.toHaveBeenCalled()
    })

    it('successfully logs in and navigates to dashboard', async () => {
        ; (IsFirstRun as any).mockResolvedValue(false)
            ; (LoginApi as any).mockResolvedValue(null)

        const wrapper = mount(Login, mountOptions)
        await new Promise(resolve => setTimeout(resolve, 0))

        const vm = wrapper.vm as any
        vm.formData.username = 'admin'
        vm.formData.password = 'password'

        await vm.handleSubmit()

        expect(LoginApi).toHaveBeenCalledWith('admin', 'password')
        expect(pushMock).toHaveBeenCalledWith('/dashboard')
    })

    it('switches to recovery mode and resets password', async () => {
        ; (IsFirstRun as any).mockResolvedValue(false)
        const wrapper = mount(Login, mountOptions)
        await new Promise(resolve => setTimeout(resolve, 0))

        const vm = wrapper.vm as any
        vm.isRecoveryMode = true
        await wrapper.vm.$nextTick()

        expect(wrapper.text()).toContain('Відновлення доступу')

        vm.formData.masterKey = 'MASTER-KEY'
        vm.formData.password = 'newpass'
        vm.formData.confirmPassword = 'newpass'

        await vm.handleSubmit()

        expect(RecoverByMasterKey).toHaveBeenCalledWith('MASTER-KEY', 'newpass')
    })
})
