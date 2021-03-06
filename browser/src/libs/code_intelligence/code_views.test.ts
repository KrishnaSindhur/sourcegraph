import { of } from 'rxjs'
import { toArray } from 'rxjs/operators'
import * as sinon from 'sinon'
import { Omit } from 'utility-types'
import { FileInfo } from './code_intelligence'
import { CodeView, toCodeViewResolver, trackCodeViews } from './code_views'

describe('code_views', () => {
    beforeEach(() => {
        document.body.innerHTML = ''
    })
    describe('trackCodeViews()', () => {
        const fileInfo: FileInfo = {
            rawRepoName: 'foo',
            filePath: '/bar.ts',
            commitID: '1',
        }
        const codeViewSpec: Omit<CodeView, 'element'> = {
            dom: {
                getCodeElementFromTarget: () => null,
                getCodeElementFromLineNumber: () => null,
                getLineElementFromLineNumber: () => null,
                getLineNumberFromCodeElement: () => 1,
            },
            resolveFileInfo: () => of(fileInfo),
        }
        it('should detect added code views from specs', async () => {
            const element = document.createElement('div')
            element.className = 'test-code-view'
            document.body.append(element)
            const selector = '.test-code-view'
            const detected = await of([{ addedNodes: [document.body], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [toCodeViewResolver(selector, codeViewSpec)],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected.map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
        })
        it('should detect added code views from resolver', async () => {
            const element = document.createElement('div')
            element.className = 'test-code-view'
            document.body.append(element)
            const selector = '.test-code-view'
            const resolveView = sinon.spy((element: HTMLElement) => ({ element, ...codeViewSpec }))
            const detected = await of([{ addedNodes: [document.body], removedNodes: [] }])
                .pipe(
                    trackCodeViews({
                        codeViewResolvers: [{ selector, resolveView }],
                    }),
                    toArray()
                )
                .toPromise()
            expect(detected.map(({ subscriptions, ...rest }) => rest)).toEqual([{ ...codeViewSpec, element }])
            sinon.assert.calledOnce(resolveView)
            sinon.assert.calledWith(resolveView, element)
        })
    })
})
