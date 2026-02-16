// Utilitários de Texto Simples e Robustos

export const toUpperCase = (text: string): string => text.toUpperCase();

export const toLowerCase = (text: string): string => text.toLowerCase();

export const formatJSON = (text: string): string => {
    try { 
        const obj = JSON.parse(text);
        return JSON.stringify(obj, null, 2); 
    } catch (e) { 
        return text; 
    }
};

export const minify = (text: string): string => {
    try { 
        return JSON.stringify(JSON.parse(text)); 
    } catch (e) { 
        // Remove quebras de linha e excesso de espaços
        return text.replace(/\s+/g, ' ').trim(); 
    }
};