import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('API Modules', () => {
    it('auth module exports expected functions', async () => {
        const authModule = await import('../api/auth')
        expect(authModule.login).toBeDefined()
        expect(authModule.register).toBeDefined()
    })

    it('question module exports expected functions', async () => {
        const questionModule = await import('../api/question')
        expect(questionModule.fetchQuestions).toBeDefined()
        expect(questionModule.createQuestion).toBeDefined()
    })

    it('paper module exports expected functions', async () => {
        const paperModule = await import('../api/paper')
        expect(paperModule.fetchPapers).toBeDefined()
        expect(paperModule.createPaper).toBeDefined()
    })

    it('exam module exports expected functions', async () => {
        const examModule = await import('../api/exam')
        expect(examModule.fetchPublishedPapers).toBeDefined()
    })

    it('document module exports expected functions', async () => {
        const documentModule = await import('../api/document')
        expect(documentModule.fetchDocumentCourses).toBeDefined()
        expect(documentModule.fetchDocumentDetail).toBeDefined()
        expect(documentModule.createDocumentCourse).toBeDefined()
    })
})
